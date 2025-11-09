import React, {useEffect, useState} from 'react';
import {createSecret, SecretOwner} from '../api/secrets';
import {listRoles, Role} from '../api/roles';

interface CreateSecretModalProps {
    open: boolean;
    onClose: () => void;
    onCreated: () => void;
}

export default function CreateSecretModal({open, onClose, onCreated}: CreateSecretModalProps) {
    const [path, setPath] = useState('');
    const [value, setValue] = useState('');
    const [users, setUsers] = useState<SecretOwner[]>([]);
    const [roles, setRoles] = useState<Role[]>([]);
    const [selectedUsers, setSelectedUsers] = useState<number[]>([]);
    const [selectedRoles, setSelectedRoles] = useState<number[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (!open) return;

        // Fetch users for selection
        async function fetchUsers() {
            setError(null);
            try {
                const resp = await fetch('/users', {
                    headers: {'Authorization': `Bearer ${localStorage.getItem('authToken')}`},
                });
                if (!resp.ok) throw new Error('Failed to fetch users');
                const data = await resp.json();
                setUsers(data);
            } catch (e: any) {
                setError('Failed to load users');
            }
        }

        // Fetch roles for selection
        async function fetchRoles() {
            setError(null);
            try {
                const data = await listRoles();
                setRoles(data);
            } catch (e: any) {
                setError('Failed to load roles');
            }
        }

        fetchUsers();
        fetchRoles();
    }, [open]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError(null);
        try {
            if (!path.startsWith('/')) throw new Error('Path must start with /');
            if (!value) throw new Error('Value is required');
            await createSecret({
                path,
                value,
                authorized_users: selectedUsers,
                authorized_roles: selectedRoles,
            });
            setPath('');
            setValue('');
            setSelectedUsers([]);
            setSelectedRoles([]);
            onCreated();
            onClose();
        } catch (e: any) {
            setError(e?.message || 'Failed to create secret');
        } finally {
            setLoading(false);
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
            <form onSubmit={handleSubmit} style={{
                background: '#fff',
                borderRadius: 10,
                padding: 32,
                minWidth: 340,
                boxShadow: '0 2px 16px rgba(0,0,0,0.12)'
            }}>
                <h3 style={{marginBottom: 18}}>Create New Secret</h3>
                <div style={{marginBottom: 12}}>
                    <label>Path</label>
                    <input type="text" value={path} onChange={e => setPath(e.target.value)} placeholder="/folder/secret"
                           style={{width: '100%', padding: 8, borderRadius: 5, border: '1px solid #ccc'}} required/>
                </div>
                <div style={{marginBottom: 12}}>
                    <label>Value</label>
                    <textarea
                        value={value}
                        onChange={e => setValue(e.target.value)}
                        placeholder="Secret value"
                        style={{
                            width: '100%',
                            padding: 8,
                            borderRadius: 5,
                            border: '1px solid #ccc',
                            minHeight: 60,
                            fontFamily: 'monospace',
                            resize: 'vertical',
                        }}
                        required
                    />
                </div>
                <div style={{marginBottom: 12}}>
                    <label>Authorized Users</label>
                    <select multiple value={selectedUsers.map(String)}
                            onChange={e => setSelectedUsers(Array.from(e.target.selectedOptions, o => Number(o.value)))}
                            style={{
                                width: '100%',
                                padding: 8,
                                borderRadius: 5,
                                border: '1px solid #ccc',
                                minHeight: 60
                            }}>
                        {users.map(u => (
                            <option key={u.id} value={u.id}>{u.username}</option>
                        ))}
                    </select>
                </div>
                <div style={{marginBottom: 12}}>
                    <label>Authorized Roles</label>
                    <select multiple value={selectedRoles.map(String)}
                            onChange={e => setSelectedRoles(Array.from(e.target.selectedOptions, o => Number(o.value)))}
                            style={{
                                width: '100%',
                                padding: 8,
                                borderRadius: 5,
                                border: '1px solid #ccc',
                                minHeight: 60
                            }}>
                        {roles.map(r => (
                            <option key={r.id} value={r.id}>{r.name}</option>
                        ))}
                    </select>
                </div>
                {error && <div style={{color: 'red', marginBottom: 10}}>{error}</div>}
                <div style={{display: 'flex', gap: 12, marginTop: 18}}>
                    <button type="button" onClick={onClose} style={{
                        flex: 1,
                        padding: 10,
                        borderRadius: 5,
                        border: '1px solid #ccc',
                        background: '#f5f5f5'
                    }} disabled={loading}>Cancel
                    </button>
                    <button type="submit" style={{
                        flex: 2,
                        padding: 10,
                        borderRadius: 5,
                        border: 'none',
                        background: '#2563eb',
                        color: '#fff',
                        fontWeight: 600
                    }} disabled={loading}>{loading ? 'Creating...' : 'Create'}</button>
                </div>
            </form>
        </div>
    );
} 