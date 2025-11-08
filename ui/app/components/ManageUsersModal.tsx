import React, {useEffect, useState} from 'react';
import {createUser, deleteUser, listUsers, User} from '../api/users';
import {assignRoleToUser, listRoles, Role, unassignRoleFromUser} from '../api/roles';

interface ManageUsersModalProps {
    open: boolean;
    onClose: () => void;
}

export default function ManageUsersModal({open, onClose}: ManageUsersModalProps) {
    const [users, setUsers] = useState<User[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [creating, setCreating] = useState(false);
    const [deletingUsernames, setDeletingUsernames] = useState<Set<string>>(new Set());
    const [currentUsername, setCurrentUsername] = useState<string | null>(null);
    const [allRoles, setAllRoles] = useState<Role[]>([]);
    const [loadingRoles, setLoadingRoles] = useState(false);
    const [showRoleSelector, setShowRoleSelector] = useState<number | null>(null);
    const [managingRoles, setManagingRoles] = useState<Set<string>>(new Set());

    useEffect(() => {
        const storedUsername = localStorage.getItem('username');
        setCurrentUsername(storedUsername);
    }, []);

    const fetchUsers = async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await listUsers();
            setUsers(data);
        } catch (e: any) {
            setError(e?.body?.message || e?.message || 'Failed to load users');
        } finally {
            setLoading(false);
        }
    };

    const fetchRoles = async () => {
        setLoadingRoles(true);
        try {
            const data = await listRoles();
            setAllRoles(data);
        } catch (e: any) {
            setError(e?.body?.message || e?.message || 'Failed to load roles');
        } finally {
            setLoadingRoles(false);
        }
    };

    useEffect(() => {
        if (open) {
            fetchUsers();
            fetchRoles();
        }
    }, [open]);

    const handleCreateUser = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!username.trim()) {
            setError('Username is required');
            return;
        }
        if (username.trim().length < 3) {
            setError('Username must be at least 3 characters');
            return;
        }
        if (!password) {
            setError('Password is required');
            return;
        }
        setCreating(true);
        setError(null);
        try {
            await createUser({username: username.trim(), password});
            setUsername('');
            setPassword('');
            await fetchUsers();
        } catch (e: any) {
            setError(e?.body?.message || e?.message || 'Failed to create user');
        } finally {
            setCreating(false);
        }
    };

    const handleDeleteUser = async (usernameToDelete: string) => {
        if (!confirm(`Are you sure you want to delete user "${usernameToDelete}"?`)) {
            return;
        }
        setDeletingUsernames(prev => new Set(prev).add(usernameToDelete));
        setError(null);
        try {
            await deleteUser(usernameToDelete);
            await fetchUsers();
        } catch (e: any) {
            setError(e?.body?.message || e?.message || 'Failed to delete user');
        } finally {
            setDeletingUsernames(prev => {
                const next = new Set(prev);
                next.delete(usernameToDelete);
                return next;
            });
        }
    };

    const handleAssignRole = async (userId: number, roleId: number) => {
        const key = `${userId}-${roleId}`;
        setManagingRoles(prev => new Set(prev).add(key));
        setError(null);
        try {
            await assignRoleToUser({user_id: userId, role_id: roleId});
            await fetchUsers();
            setShowRoleSelector(null);
        } catch (e: any) {
            setError(e?.body?.message || e?.message || 'Failed to assign role');
        } finally {
            setManagingRoles(prev => {
                const next = new Set(prev);
                next.delete(key);
                return next;
            });
        }
    };

    const handleUnassignRole = async (userId: number, roleId: number) => {
        const key = `${userId}-${roleId}`;
        setManagingRoles(prev => new Set(prev).add(key));
        setError(null);
        try {
            await unassignRoleFromUser({user_id: userId, role_id: roleId});
            await fetchUsers();
        } catch (e: any) {
            setError(e?.body?.message || e?.message || 'Failed to unassign role');
        } finally {
            setManagingRoles(prev => {
                const next = new Set(prev);
                next.delete(key);
                return next;
            });
        }
    };

    const getAvailableRoles = (user: User): Role[] => {
        const userRoleIds = new Set(user.roles?.map(r => r.id) || []);
        return allRoles.filter(role => !userRoleIds.has(role.id));
    };

    useEffect(() => {
        if (open) {
            console.log('ManageUsersModal opened');
        }
    }, [open]);

    // Close role selector when clicking outside
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (showRoleSelector !== null) {
                const target = event.target as HTMLElement;
                if (!target.closest('[data-role-selector]')) {
                    setShowRoleSelector(null);
                }
            }
        };
        if (showRoleSelector !== null) {
            document.addEventListener('mousedown', handleClickOutside);
            return () => {
                document.removeEventListener('mousedown', handleClickOutside);
            };
        }
    }, [showRoleSelector]);

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
                minWidth: 600,
                maxWidth: 900,
                maxHeight: '90vh',
                boxShadow: '0 2px 16px rgba(0,0,0,0.12)',
                display: 'flex',
                flexDirection: 'column',
                overflow: 'hidden'
            }}>
                <h3 style={{marginBottom: 18}}>Manage Users</h3>

                {/* Create User Form */}
                <form onSubmit={handleCreateUser} style={{
                    marginBottom: 24,
                    padding: 16,
                    background: '#f9fafb',
                    borderRadius: 8,
                    border: '1px solid #e5e7eb'
                }}>
                    <h4 style={{marginBottom: 12, fontSize: 16}}>Create New User</h4>
                    <div style={{display: 'flex', flexDirection: 'column', gap: 12, marginBottom: 12}}>
                        <div>
                            <label style={{display: 'block', marginBottom: 4, fontSize: 14}}>Username</label>
                            <input
                                type="text"
                                value={username}
                                onChange={e => setUsername(e.target.value)}
                                placeholder="Username (min 3 characters)"
                                style={{
                                    width: '100%',
                                    padding: 8,
                                    borderRadius: 5,
                                    border: '1px solid #ccc'
                                }}
                                required
                                minLength={3}
                            />
                        </div>
                        <div>
                            <label style={{display: 'block', marginBottom: 4, fontSize: 14}}>Password</label>
                            <input
                                type="password"
                                value={password}
                                onChange={e => setPassword(e.target.value)}
                                placeholder="Password"
                                style={{
                                    width: '100%',
                                    padding: 8,
                                    borderRadius: 5,
                                    border: '1px solid #ccc'
                                }}
                                required
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
                        {creating ? 'Creating...' : 'Create User'}
                    </button>
                </form>

                {error && <div style={{
                    color: 'red',
                    marginBottom: 12,
                    padding: 8,
                    background: '#fee',
                    borderRadius: 5
                }}>{error}</div>}

                {/* Users List */}
                <div style={{flex: 1, overflowY: 'auto', marginBottom: 18}}>
                    <h4 style={{marginBottom: 12, fontSize: 16}}>Existing Users</h4>
                    {loading ? (
                        <div>Loading users...</div>
                    ) : users.length === 0 ? (
                        <div style={{color: '#666', fontStyle: 'italic'}}>No users found</div>
                    ) : (
                        <div style={{display: 'flex', flexDirection: 'column', gap: 8}}>
                            {users.map(user => (
                                <div
                                    key={user.id}
                                    style={{
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'space-between',
                                        padding: 12,
                                        borderRadius: 6,
                                        border: '1px solid #e5e7eb',
                                        background: '#fff',
                                        gap: 12
                                    }}
                                >
                                    <div style={{flex: 1, minWidth: 0, position: 'relative'}}>
                                        <div style={{display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4}}>
                                            <span style={{fontWeight: 500}}>{user.username}</span>
                                            {currentUsername && user.username === currentUsername && (
                                                <span style={{
                                                    fontSize: 11,
                                                    padding: '2px 6px',
                                                    borderRadius: 4,
                                                    background: '#dbeafe',
                                                    color: '#1e40af',
                                                    fontWeight: 600
                                                }}>
                                                    You
                                                </span>
                                            )}
                                            {user.id === 1 && (
                                                <span style={{
                                                    fontSize: 11,
                                                    padding: '2px 6px',
                                                    borderRadius: 4,
                                                    background: '#f3f4f6',
                                                    color: '#374151',
                                                    fontWeight: 600
                                                }}>
                                                    System
                                                </span>
                                            )}
                                        </div>
                                        <div style={{
                                            fontSize: 12,
                                            color: '#666',
                                            display: 'flex',
                                            gap: 12,
                                            flexWrap: 'wrap',
                                            marginBottom: 6
                                        }}>
                                            <span>ID: {user.id}</span>
                                            {user.created_at && (
                                                <span>Created: {new Date(user.created_at).toLocaleString()}</span>
                                            )}
                                            {user.updated_at && (
                                                <span>Updated: {new Date(user.updated_at).toLocaleString()}</span>
                                            )}
                                        </div>
                                        <div style={{
                                            display: 'flex',
                                            flexWrap: 'wrap',
                                            gap: 6,
                                            alignItems: 'center',
                                            marginTop: 6
                                        }}>
                                            <span style={{fontSize: 12, color: '#666', marginRight: 4}}>Roles:</span>
                                            {user.roles && user.roles.length > 0 ? (
                                                user.roles.map((role: Role) => {
                                                    const managingKey = `${user.id}-${role.id}`;
                                                    const isManaging = managingRoles.has(managingKey);
                                                    return (
                                                        <div
                                                            key={role.id}
                                                            style={{
                                                                display: 'flex',
                                                                alignItems: 'center',
                                                                gap: 4,
                                                                padding: '4px 8px',
                                                                borderRadius: 4,
                                                                background: '#f9fafb',
                                                                border: '1px solid #e5e7eb',
                                                                fontSize: 12
                                                            }}
                                                        >
                                                            <div
                                                                style={{
                                                                    width: 12,
                                                                    height: 12,
                                                                    borderRadius: 3,
                                                                    background: role.color,
                                                                    border: '1px solid #ccc',
                                                                    flexShrink: 0
                                                                }}
                                                            />
                                                            <span style={{fontWeight: 500}}>{role.name}</span>
                                                            {role.admin && (
                                                                <span style={{
                                                                    fontSize: 10,
                                                                    padding: '1px 4px',
                                                                    borderRadius: 3,
                                                                    background: '#fef3c7',
                                                                    color: '#92400e',
                                                                    fontWeight: 600
                                                                }}>
                                                                    ADMIN
                                                                </span>
                                                            )}
                                                            <button
                                                                type="button"
                                                                onClick={() => handleUnassignRole(user.id, role.id)}
                                                                disabled={isManaging}
                                                                style={{
                                                                    marginLeft: 4,
                                                                    padding: '2px 4px',
                                                                    border: 'none',
                                                                    background: 'transparent',
                                                                    color: '#dc2626',
                                                                    cursor: isManaging ? 'not-allowed' : 'pointer',
                                                                    fontSize: 14,
                                                                    lineHeight: 1,
                                                                    display: 'flex',
                                                                    alignItems: 'center',
                                                                    justifyContent: 'center',
                                                                    opacity: isManaging ? 0.5 : 1
                                                                }}
                                                                title="Remove role"
                                                            >
                                                                ×
                                                            </button>
                                                        </div>
                                                    );
                                                })
                                            ) : (
                                                <span style={{fontSize: 12, color: '#999', fontStyle: 'italic'}}>No roles assigned</span>
                                            )}
                                            <div style={{position: 'relative', display: 'inline-block'}}
                                                 data-role-selector>
                                                <button
                                                    type="button"
                                                    onClick={() => setShowRoleSelector(showRoleSelector === user.id ? null : user.id)}
                                                    disabled={loadingRoles || managingRoles.size > 0}
                                                    style={{
                                                        padding: '4px 8px',
                                                        borderRadius: 4,
                                                        border: '1px solid #2563eb',
                                                        background: '#fff',
                                                        color: '#2563eb',
                                                        fontSize: 12,
                                                        fontWeight: 500,
                                                        cursor: (loadingRoles || managingRoles.size > 0) ? 'not-allowed' : 'pointer',
                                                        opacity: (loadingRoles || managingRoles.size > 0) ? 0.5 : 1
                                                    }}
                                                >
                                                    + Add Role
                                                </button>
                                                {showRoleSelector === user.id && (
                                                    <div style={{
                                                        position: 'absolute',
                                                        top: '100%',
                                                        left: 0,
                                                        marginTop: 4,
                                                        padding: 8,
                                                        background: '#fff',
                                                        border: '1px solid #e5e7eb',
                                                        borderRadius: 6,
                                                        boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                                                        zIndex: 1001,
                                                        maxHeight: 200,
                                                        overflowY: 'auto',
                                                        minWidth: 200
                                                    }}>
                                                        {getAvailableRoles(user).length === 0 ? (
                                                            <div style={{fontSize: 12, color: '#666', padding: 8}}>No
                                                                available roles</div>
                                                        ) : (
                                                            getAvailableRoles(user).map(role => {
                                                                const managingKey = `${user.id}-${role.id}`;
                                                                const isManaging = managingRoles.has(managingKey);
                                                                return (
                                                                    <button
                                                                        key={role.id}
                                                                        type="button"
                                                                        onClick={() => handleAssignRole(user.id, role.id)}
                                                                        disabled={isManaging}
                                                                        style={{
                                                                            width: '100%',
                                                                            display: 'flex',
                                                                            alignItems: 'center',
                                                                            gap: 8,
                                                                            padding: '6px 8px',
                                                                            border: 'none',
                                                                            background: 'transparent',
                                                                            cursor: isManaging ? 'not-allowed' : 'pointer',
                                                                            fontSize: 12,
                                                                            textAlign: 'left',
                                                                            opacity: isManaging ? 0.5 : 1
                                                                        }}
                                                                        onMouseEnter={(e) => {
                                                                            if (!isManaging) {
                                                                                e.currentTarget.style.background = '#f9fafb';
                                                                            }
                                                                        }}
                                                                        onMouseLeave={(e) => {
                                                                            e.currentTarget.style.background = 'transparent';
                                                                        }}
                                                                    >
                                                                        <div
                                                                            style={{
                                                                                width: 12,
                                                                                height: 12,
                                                                                borderRadius: 3,
                                                                                background: role.color,
                                                                                border: '1px solid #ccc',
                                                                                flexShrink: 0
                                                                            }}
                                                                        />
                                                                        <span
                                                                            style={{fontWeight: 500}}>{role.name}</span>
                                                                        {role.admin && (
                                                                            <span style={{
                                                                                fontSize: 10,
                                                                                padding: '1px 4px',
                                                                                borderRadius: 3,
                                                                                background: '#fef3c7',
                                                                                color: '#92400e',
                                                                                fontWeight: 600,
                                                                                marginLeft: 'auto'
                                                                            }}>
                                                                            ADMIN
                                                                        </span>
                                                                        )}
                                                                    </button>
                                                                );
                                                            })
                                                        )}
                                                    </div>
                                                )}
                                            </div>
                                        </div>
                                    </div>
                                    <button
                                        type="button"
                                        onClick={() => handleDeleteUser(user.username)}
                                        disabled={deletingUsernames.has(user.username)}
                                        style={{
                                            padding: '6px 12px',
                                            borderRadius: 5,
                                            border: '1px solid #dc2626',
                                            background: deletingUsernames.has(user.username) ? '#fca5a5' : '#ef4444',
                                            color: '#fff',
                                            fontWeight: 500,
                                            cursor: deletingUsernames.has(user.username) ? 'not-allowed' : 'pointer',
                                            fontSize: 14
                                        }}
                                    >
                                        {deletingUsernames.has(user.username) ? 'Deleting...' : 'Delete'}
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

