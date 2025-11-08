import React, {useEffect, useState} from 'react';
import {createRole, deleteRole, listRoles, Role} from '../api/roles';

interface ManageRolesModalProps {
    open: boolean;
    onClose: () => void;
}

export default function ManageRolesModal({open, onClose}: ManageRolesModalProps) {
    const [roles, setRoles] = useState<Role[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [name, setName] = useState('');
    const [color, setColor] = useState('#2563eb');
    const [creating, setCreating] = useState(false);
    const [deletingIds, setDeletingIds] = useState<Set<number>>(new Set());

    const fetchRoles = async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await listRoles();
            setRoles(data);
        } catch (e: any) {
            setError(e?.body?.message || e?.message || 'Failed to load roles');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (open) {
            fetchRoles();
        }
    }, [open]);

    const handleCreateRole = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!name.trim()) {
            setError('Role name is required');
            return;
        }
        setCreating(true);
        setError(null);
        try {
            await createRole({name: name.trim(), color});
            setName('');
            setColor('#2563eb');
            await fetchRoles();
        } catch (e: any) {
            setError(e?.body?.message || e?.message || 'Failed to create role');
        } finally {
            setCreating(false);
        }
    };

    const handleDeleteRole = async (roleId: number) => {
        if (!confirm(`Are you sure you want to delete this role?`)) {
            return;
        }
        setDeletingIds(prev => new Set(prev).add(roleId));
        setError(null);
        try {
            await deleteRole(roleId);
            await fetchRoles();
        } catch (e: any) {
            setError(e?.body?.message || e?.message || 'Failed to delete role');
        } finally {
            setDeletingIds(prev => {
                const next = new Set(prev);
                next.delete(roleId);
                return next;
            });
        }
    };

    if (!open) return null;

    return (
        <div style={{
            position: 'fixed',
            top: 0,
            left: 0,
            width: '100vw',
            height: '100vh',
            background: 'rgba(0,0,0,0.18)',
            zIndex: 1000,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center'
        }}>
            <div style={{
                background: '#fff',
                borderRadius: 10,
                padding: 32,
                minWidth: 500,
                maxWidth: 700,
                maxHeight: '90vh',
                boxShadow: '0 2px 16px rgba(0,0,0,0.12)',
                display: 'flex',
                flexDirection: 'column',
                overflow: 'hidden'
            }}>
                <h3 style={{marginBottom: 18}}>Manage Roles</h3>

                {/* Create Role Form */}
                <form onSubmit={handleCreateRole} style={{
                    marginBottom: 24,
                    padding: 16,
                    background: '#f9fafb',
                    borderRadius: 8,
                    border: '1px solid #e5e7eb'
                }}>
                    <h4 style={{marginBottom: 12, fontSize: 16}}>Create New Role</h4>
                    <div style={{display: 'flex', gap: 12, marginBottom: 12}}>
                        <div style={{flex: 2}}>
                            <label style={{display: 'block', marginBottom: 4, fontSize: 14}}>Name</label>
                            <input
                                type="text"
                                value={name}
                                onChange={e => setName(e.target.value)}
                                placeholder="Role name"
                                style={{
                                    width: '100%',
                                    padding: 8,
                                    borderRadius: 5,
                                    border: '1px solid #ccc'
                                }}
                                required
                            />
                        </div>
                        <div style={{flex: 1}}>
                            <label style={{display: 'block', marginBottom: 4, fontSize: 14}}>Color</label>
                            <input
                                type="color"
                                value={color}
                                onChange={e => setColor(e.target.value)}
                                style={{
                                    width: '100%',
                                    padding: 4,
                                    borderRadius: 5,
                                    border: '1px solid #ccc',
                                    height: 36
                                }}
                            />
                        </div>
                    </div>
                    <button
                        type="submit"
                        disabled={creating}
                        style={{
                            padding: '8px 16px',
                            borderRadius: 5,
                            border: 'none',
                            background: '#2563eb',
                            color: '#fff',
                            fontWeight: 600,
                            cursor: creating ? 'not-allowed' : 'pointer'
                        }}
                    >
                        {creating ? 'Creating...' : 'Create Role'}
                    </button>
                </form>

                {error && <div style={{
                    color: 'red',
                    marginBottom: 12,
                    padding: 8,
                    background: '#fee',
                    borderRadius: 5
                }}>{error}</div>}

                {/* Roles List */}
                <div style={{flex: 1, overflowY: 'auto', marginBottom: 18}}>
                    <h4 style={{marginBottom: 12, fontSize: 16}}>Existing Roles</h4>
                    {loading ? (
                        <div>Loading roles...</div>
                    ) : roles.length === 0 ? (
                        <div style={{color: '#666', fontStyle: 'italic'}}>No roles found</div>
                    ) : (
                        <div style={{display: 'flex', flexDirection: 'column', gap: 8}}>
                            {roles.map(role => (
                                <div
                                    key={role.id}
                                    style={{
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'space-between',
                                        padding: 12,
                                        borderRadius: 6,
                                        border: '1px solid #e5e7eb',
                                        background: '#fff'
                                    }}
                                >
                                    <div style={{display: 'flex', alignItems: 'center', gap: 12, flex: 1}}>
                                        <div
                                            style={{
                                                width: 24,
                                                height: 24,
                                                borderRadius: 4,
                                                background: role.color,
                                                border: '1px solid #ccc'
                                            }}
                                        />
                                        <div style={{flex: 1}}>
                                            <div style={{
                                                display: 'flex',
                                                alignItems: 'center',
                                                gap: 8,
                                                marginBottom: 4
                                            }}>
                                                <span style={{fontWeight: 500}}>{role.name}</span>
                                                {role.admin && (
                                                    <span style={{
                                                        fontSize: 11,
                                                        padding: '2px 6px',
                                                        borderRadius: 4,
                                                        background: '#fef3c7',
                                                        color: '#92400e',
                                                        fontWeight: 600
                                                    }}>
                                                        ADMIN
                                                    </span>
                                                )}
                                            </div>
                                            <div style={{
                                                fontSize: 12,
                                                color: '#666',
                                                display: 'flex',
                                                gap: 12,
                                                flexWrap: 'wrap'
                                            }}>
                                                <span>ID: {role.id}</span>
                                                {role.created_at && (
                                                    <span>Created: {new Date(role.created_at).toLocaleString()}</span>
                                                )}
                                            </div>
                                        </div>
                                    </div>
                                    <button
                                        type="button"
                                        onClick={() => handleDeleteRole(role.id)}
                                        disabled={deletingIds.has(role.id)}
                                        style={{
                                            padding: '6px 12px',
                                            borderRadius: 5,
                                            border: '1px solid #dc2626',
                                            background: deletingIds.has(role.id) ? '#fca5a5' : '#ef4444',
                                            color: '#fff',
                                            fontWeight: 500,
                                            cursor: deletingIds.has(role.id) ? 'not-allowed' : 'pointer',
                                            fontSize: 14
                                        }}
                                    >
                                        {deletingIds.has(role.id) ? 'Deleting...' : 'Delete'}
                                    </button>
                                </div>
                            ))}
                        </div>
                    )}
                </div>

                <div style={{display: 'flex', gap: 12}}>
                    <button
                        type="button"
                        onClick={onClose}
                        style={{
                            flex: 1,
                            padding: 10,
                            borderRadius: 5,
                            border: '1px solid #ccc',
                            background: '#f5f5f5',
                            cursor: 'pointer'
                        }}
                    >
                        Close
                    </button>
                </div>
            </div>
        </div>
    );
}

