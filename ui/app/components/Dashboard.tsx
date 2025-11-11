'use client'

import React, {useEffect, useState} from 'react';
import {getSecret, listSecrets, SecretInfoSummary, updateSecret} from '../api/secrets';
import {listUsers, User} from '../api/users';
import {listRoles, Role} from '../api/roles';
import {FolderNode} from '../types';
import CreateSecretModal from './CreateSecretModal';
import ManageRolesModal from './ManageRolesModal';
import ManageUsersModal from './ManageUsersModal';

function buildFolderTree(secrets: SecretInfoSummary[]): FolderNode {
    const root: FolderNode = {name: '', path: '', children: [], secrets: [], isFolder: true};
    for (const secret of secrets) {
        const parts = secret.path.split('/').filter(Boolean);
        let node = root;
        let currentPath = '';
        for (let i = 0; i < parts.length - 1; i++) {
            currentPath += '/' + parts[i];
            let child = node.children.find(c => c.name === parts[i] && c.isFolder);
            if (!child) {
                child = {name: parts[i], path: currentPath, children: [], secrets: [], isFolder: true};
                node.children.push(child);
            }
            node = child;
        }
        node.secrets.push(secret);
    }
    return root;
}

function FolderView({node, onSelectSecret}: { node: FolderNode; onSelectSecret: (secret: SecretInfoSummary) => void }) {
    return (
        <div style={{marginLeft: 16}}>
            {node.children.map(child => (
                <div key={child.path}>
                    <div style={{fontWeight: 'bold', marginTop: 8}}>{child.name}/</div>
                    <FolderView node={child} onSelectSecret={onSelectSecret}/>
                </div>
            ))}
            {node.secrets.map(secret => (
                <div
                    key={secret.path}
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        cursor: 'pointer',
                        padding: '8px 6px',
                        margin: '4px 0',
                        borderRadius: 6,
                        border: '1px solid #e0e7ef',
                        boxShadow: '0 1px 3px rgba(0,0,0,0.03)',
                        background: '#fff',
                        transition: 'background 0.15s, box-shadow 0.15s',
                    }}
                    onMouseOver={e => {
                        (e.currentTarget as HTMLDivElement).style.background = '#f0f6ff';
                        (e.currentTarget as HTMLDivElement).style.boxShadow = '0 2px 8px rgba(37,99,235,0.08)';
                    }}
                    onMouseOut={e => {
                        (e.currentTarget as HTMLDivElement).style.background = '#fff';
                        (e.currentTarget as HTMLDivElement).style.boxShadow = '0 1px 3px rgba(0,0,0,0.03)';
                    }}
                    onClick={() => onSelectSecret(secret)}
                >
                    <span style={{fontWeight: 500, minWidth: 120}}>{secret.path.split('/').pop()}</span>
                    <div style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 4,
                        flexWrap: 'wrap',
                        flex: 1,
                        marginLeft: 8
                    }}>
                        {secret.owner && (
                            <div
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: 4,
                                    padding: '2px 6px',
                                    borderRadius: 4,
                                    background: '#fef3c7',
                                    border: '1px solid #fbbf24',
                                    fontSize: 11,
                                }}
                                title={`Owner: ${secret.owner.username}`}
                            >
                                <span style={{color: '#92400e', fontSize: 10}}>👑</span>
                                <span style={{fontWeight: 600, color: '#78350f'}}>{secret.owner.username}</span>
                            </div>
                        )}
                        {secret.roles && secret.roles.length > 0 && (
                            <>
                                {secret.roles.map(role => (
                                    <div
                                        key={role.id}
                                        style={{
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: 4,
                                            padding: '2px 6px',
                                            borderRadius: 4,
                                            background: '#f9fafb',
                                            border: '1px solid #e5e7eb',
                                            fontSize: 11,
                                        }}
                                        title={`Role: ${role.name}`}
                                    >
                                        <div
                                            style={{
                                                width: 10,
                                                height: 10,
                                                borderRadius: 2,
                                                background: role.color,
                                                border: '1px solid #ccc',
                                                flexShrink: 0
                                            }}
                                        />
                                        <span style={{fontWeight: 500, color: '#374151'}}>{role.name}</span>
                                    </div>
                                ))}
                            </>
                        )}
                        {secret.users && secret.users.length > 0 && (
                            <>
                                {secret.users.map(user => (
                                    <div
                                        key={user.id}
                                        style={{
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: 4,
                                            padding: '2px 6px',
                                            borderRadius: 4,
                                            background: '#f0f9ff',
                                            border: '1px solid #bae6fd',
                                            fontSize: 11,
                                        }}
                                        title={`User: ${user.username}`}
                                    >
                                        <span style={{color: '#0369a1', fontSize: 10}}>👤</span>
                                        <span style={{fontWeight: 500, color: '#0c4a6e'}}>{user.username}</span>
                                    </div>
                                ))}
                            </>
                        )}
                    </div>
                    <span style={{
                        color: '#888',
                        fontSize: 12,
                        marginLeft: 8,
                        whiteSpace: 'nowrap'
                    }}>Last updated: {new Date(secret.updated_at).toLocaleString()}</span>
                </div>
            ))}
        </div>
    );
}

export default function Dashboard() {
    const [secrets, setSecrets] = useState<SecretInfoSummary[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [selectedSecret, setSelectedSecret] = useState<SecretInfoSummary | null>(null);
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [showManageRolesModal, setShowManageRolesModal] = useState(false);
    const [showManageUsersModal, setShowManageUsersModal] = useState(false);
    const [secretInfo, setSecretInfo] = useState<any>(null);
    const [secretInfoLoading, setSecretInfoLoading] = useState(false);
    const [secretInfoError, setSecretInfoError] = useState<string | null>(null);
    const [showValue, setShowValue] = useState(false);
    const [secretValue, setSecretValue] = useState<string | null>(null);
    const [secretValueLoading, setSecretValueLoading] = useState(false);
    const [secretValueError, setSecretValueError] = useState<string | null>(null);
    const [secretValueCopied, setSecretValueCopied] = useState(false);
    const [isEditing, setIsEditing] = useState(false);
    const [editValue, setEditValue] = useState('');
    const [editSelectedUsers, setEditSelectedUsers] = useState<number[]>([]);
    const [editSelectedRoles, setEditSelectedRoles] = useState<number[]>([]);
    const [editUsers, setEditUsers] = useState<User[]>([]);
    const [editRoles, setEditRoles] = useState<Role[]>([]);
    const [editLoading, setEditLoading] = useState(false);
    const [editError, setEditError] = useState<string | null>(null);
    const [showUserSelector, setShowUserSelector] = useState(false);
    const [showRoleSelector, setShowRoleSelector] = useState(false);

    const fetchSecrets = async () => {
        setLoading(true);
        setError(null);
        try {
            const secrets = await listSecrets();
            setSecrets(secrets);
        } catch (err: any) {
            setError(err?.body || 'Failed to fetch secrets');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchSecrets();
    }, []);

    useEffect(() => {
        if (!selectedSecret) {
            setSecretInfo(null);
            setSecretInfoError(null);
            setShowValue(false);
            setSecretValue(null);
            setSecretValueError(null);
            setIsEditing(false);
            return;
        }
        setSecretInfoLoading(true);
        setSecretInfoError(null);
        setSecretInfo(null);
        setShowValue(false);
        setSecretValue(null);
        setSecretValueError(null);
        setIsEditing(false);
        getSecret(btoa(selectedSecret.path))
            .then(info => setSecretInfo(info))
            .catch(e => setSecretInfoError(e?.body || 'Failed to fetch secret info'))
            .finally(() => setSecretInfoLoading(false));
    }, [selectedSecret]);

    useEffect(() => {
        if (!isEditing || !selectedSecret) return;

        async function fetchUsers() {
            try {
                const data = await listUsers();
                setEditUsers(data);
            } catch (e: any) {
                setEditError('Failed to load users');
            }
        }

        async function fetchRoles() {
            try {
                const data = await listRoles();
                setEditRoles(data);
            } catch (e: any) {
                setEditError('Failed to load roles');
            }
        }

        fetchUsers();
        fetchRoles();
    }, [isEditing, selectedSecret]);

    // Pre-populate form when entering edit mode
    useEffect(() => {
        if (isEditing && secretInfo && secretValue !== null) {
            setEditValue(secretValue);
            setEditSelectedUsers(secretInfo.authorized_users?.map((u: any) => u.id) || []);
            setEditSelectedRoles(secretInfo.authorized_roles?.map((r: any) => r.id) || []);
        }
    }, [isEditing, secretInfo, secretValue]);

    // Close selectors when clicking outside
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (showUserSelector || showRoleSelector) {
                const target = event.target as HTMLElement;
                if (!target.closest('[data-user-selector]') && !target.closest('[data-role-selector]')) {
                    setShowUserSelector(false);
                    setShowRoleSelector(false);
                }
            }
        };
        if (showUserSelector || showRoleSelector) {
            document.addEventListener('mousedown', handleClickOutside);
            return () => {
                document.removeEventListener('mousedown', handleClickOutside);
            };
        }
    }, [showUserSelector, showRoleSelector]);

    const handleShowValue = async () => {
        if (!selectedSecret) return;
        setSecretValueLoading(true);
        setSecretValueError(null);
        setShowValue(true);
        try {
            // Fetch the value using the get secret value endpoint
            const encodedPath = btoa(selectedSecret.path);
            console.log("encodedPath:", encodedPath)
            const headers: Record<string, string> = {
                'Content-Type': 'application/json',
                ...(localStorage.getItem('authToken') ? {'Authorization': `Bearer ${localStorage.getItem('authToken')}`} : {}),
            };
            const resp = await fetch(`/secrets/${encodedPath}/value`, {
                method: 'GET',
                headers,
                credentials: 'include',
            });
            if (!resp.ok) {
                let errorBody: any = null;
                try {
                    errorBody = await resp.text();
                    try {
                        errorBody = JSON.parse(errorBody);
                    } catch {
                    }
                } catch {
                }
                throw {status: resp.status, body: errorBody};
            }
            const data = await resp.json();
            setSecretValue(data.value);
        } catch (e: any) {
            setSecretValueError(e?.body || 'Failed to fetch secret value');
        } finally {
            setSecretValueLoading(false);
        }
    };

    const handleCopySecretValue = async () => {
        if (secretValue !== null) {
            try {
                await navigator.clipboard.writeText(secretValue);
                setSecretValueCopied(true);
                setTimeout(() => setSecretValueCopied(false), 1500);
            } catch {
            }
        }
    };

    const handleStartEdit = async () => {
        if (!selectedSecret || !secretInfo) return;
        // Need to fetch the value first if not already loaded
        if (!secretValue) {
            setSecretValueLoading(true);
            setSecretValueError(null);
            try {
                const encodedPath = btoa(selectedSecret.path);
                const headers: Record<string, string> = {
                    'Content-Type': 'application/json',
                    ...(localStorage.getItem('authToken') ? {'Authorization': `Bearer ${localStorage.getItem('authToken')}`} : {}),
                };
                const resp = await fetch(`/secrets/${encodedPath}/value`, {
                    method: 'GET',
                    headers,
                    credentials: 'include',
                });
                if (!resp.ok) {
                    let errorBody: any = null;
                    try {
                        errorBody = await resp.text();
                        try {
                            errorBody = JSON.parse(errorBody);
                        } catch {
                        }
                    } catch {
                    }
                    throw {status: resp.status, body: errorBody};
                }
                const data = await resp.json();
                setSecretValue(data.value);
            } catch (e: any) {
                setSecretValueError(e?.body || 'Failed to fetch secret value');
                setSecretValueLoading(false);
                return;
            } finally {
                setSecretValueLoading(false);
            }
        }
        setIsEditing(true);
        setEditError(null);
    };

    const handleCancelEdit = () => {
        setIsEditing(false);
        setEditError(null);
        setEditValue('');
        setEditSelectedUsers([]);
        setEditSelectedRoles([]);
        setShowUserSelector(false);
        setShowRoleSelector(false);
    };

    const handleRemoveUser = (userId: number) => {
        setEditSelectedUsers(prev => prev.filter(id => id !== userId));
    };

    const handleAddUser = (userId: number) => {
        if (!editSelectedUsers.includes(userId)) {
            setEditSelectedUsers(prev => [...prev, userId]);
        }
        setShowUserSelector(false);
    };

    const handleRemoveRole = (roleId: number) => {
        setEditSelectedRoles(prev => prev.filter(id => id !== roleId));
    };

    const handleAddRole = (roleId: number) => {
        if (!editSelectedRoles.includes(roleId)) {
            setEditSelectedRoles(prev => [...prev, roleId]);
        }
        setShowRoleSelector(false);
    };

    const getAvailableUsers = (): User[] => {
        return editUsers.filter(u => !editSelectedUsers.includes(u.id));
    };

    const getAvailableRoles = (): Role[] => {
        return editRoles.filter(r => !editSelectedRoles.includes(r.id));
    };

    const handleSubmitEdit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!selectedSecret) return;
        setEditLoading(true);
        setEditError(null);
        try {
            await updateSecret({
                path: selectedSecret.path,
                value: editValue,
                authorized_users: editSelectedUsers,
                authorized_roles: editSelectedRoles,
            });
            setIsEditing(false);
            // Refresh secret info and list
            await fetchSecrets();
            if (selectedSecret) {
                const info = await getSecret(btoa(selectedSecret.path));
                setSecretInfo(info);
            }
        } catch (e: any) {
            setEditError(e?.body?.message || e?.message || 'Failed to update secret');
        } finally {
            setEditLoading(false);
        }
    };

    const root = buildFolderTree(secrets);

    return (
        <div style={{padding: 24}}>
            <div style={{display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16}}>
                <h2>Secrets Dashboard</h2>
                <div style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'flex-end',
                    marginBottom: 16,
                    gap: 12
                }}>
                    <button type="button" style={{padding: '8px 16px', fontSize: 16}}
                            onClick={() => setShowCreateModal(true)}>
                        + New Secret
                    </button>
                    <button type="button" style={{padding: '8px 16px', fontSize: 16}}
                            onClick={() => setShowManageRolesModal(true)}>
                        Manage roles
                    </button>
                    <button
                        type="button"
                        style={{padding: '8px 16px', fontSize: 16}}
                        onClick={() => setShowManageUsersModal(true)}
                    >
                        Manage users
                    </button>
                </div>
            </div>
            {loading ? (
                <div>Loading secrets...</div>
            ) : error ? (
                <div style={{color: 'red'}}>{error}</div>
            ) : (
                <div style={{border: '1px solid #eee', borderRadius: 8, padding: 16, background: '#fafbfc'}}>
                    <FolderView node={root} onSelectSecret={setSelectedSecret}/>
                </div>
            )}
            {selectedSecret && (
                <div style={{
                    marginTop: 32,
                    border: '1px solid #ddd',
                    borderRadius: 8,
                    padding: 24,
                    background: '#fff',
                    minWidth: 340
                }}>
                    <h3>Secret Details: {selectedSecret.path}</h3>
                    {secretInfoLoading ? (
                        <div>Loading secret info...</div>
                    ) : secretInfoError ? (
                        <div style={{color: 'red'}}>{secretInfoError}</div>
                    ) : secretInfo ? (
                        <>
                            {!isEditing ? (
                                <>
                                    <div><b>Owner:</b> {secretInfo.owner?.username}</div>
                                    <div><b>Created:</b> {new Date(secretInfo.created_at).toLocaleString()}</div>
                                    <div><b>Last Updated:</b> {new Date(secretInfo.updated_at).toLocaleString()}</div>
                                    <div style={{marginTop: 12}}><b>Authorized
                                        Users:</b> {secretInfo.authorized_users && secretInfo.authorized_users.length > 0 ? secretInfo.authorized_users.map((u: any) => u.username).join(', ') : 'None'}
                                    </div>
                                    <div><b>Authorized
                                        Roles:</b> {secretInfo.authorized_roles && secretInfo.authorized_roles.length > 0 ? secretInfo.authorized_roles.map((r: any) => r.name).join(', ') : 'None'}
                                    </div>
                                    <div style={{marginTop: 18}}>
                                        <button onClick={handleShowValue} disabled={secretValueLoading} style={{
                                            padding: '6px 16px',
                                            fontSize: 15,
                                            borderRadius: 5,
                                            border: '1px solid #ccc',
                                            background: '#f5f5f5',
                                            marginRight: 10
                                        }}>
                                            {secretValueLoading ? 'Loading...' : (showValue ? 'Refresh Value' : 'Show Value')}
                                        </button>
                                        {showValue && (
                                            <span
                                                style={{marginLeft: 8, display: 'flex', alignItems: 'center', gap: 8}}>
                                                {secretValueError ? (
                                                    <span style={{color: 'red'}}>{secretValueError}</span>
                                                ) : secretValue !== null ? (
                                                    <>
                                                        <textarea
                                                            value={secretValue}
                                                            readOnly
                                                            style={{
                                                                fontFamily: 'monospace',
                                                                background: '#f5f5f5',
                                                                padding: '8px',
                                                                borderRadius: 4,
                                                                minWidth: 260,
                                                                minHeight: 60,
                                                                resize: 'vertical',
                                                                border: '1px solid #ccc',
                                                                marginTop: 4,
                                                                display: 'block',
                                                            }}
                                                        />
                                                        <button
                                                            type="button"
                                                            onClick={handleCopySecretValue}
                                                            style={{
                                                                marginLeft: 6,
                                                                padding: '6px 12px',
                                                                borderRadius: 4,
                                                                border: '1px solid #bbb',
                                                                background: secretValueCopied ? '#d1fae5' : '#f5f5f5',
                                                                color: '#222',
                                                                fontWeight: 500,
                                                                cursor: 'pointer',
                                                                transition: 'background 0.18s',
                                                            }}
                                                            disabled={secretValueCopied}
                                                            title="Copy to clipboard"
                                                        >
                                                            {secretValueCopied ? 'Copied!' : 'Copy'}
                                                        </button>
                                                    </>
                                                ) : null}
                                            </span>
                                        )}
                                    </div>
                                    <div style={{marginTop: 18}}>
                                        <button onClick={handleStartEdit} style={{
                                            padding: '6px 16px',
                                            fontSize: 15,
                                            borderRadius: 5,
                                            border: '1px solid #2563eb',
                                            background: '#2563eb',
                                            color: '#fff',
                                            fontWeight: 500,
                                            cursor: 'pointer',
                                        }}>
                                            Edit Secret
                                        </button>
                                    </div>
                                </>
                            ) : (
                                <form onSubmit={handleSubmitEdit} style={{marginTop: 12}}>
                                    <div style={{marginBottom: 12}}>
                                        <label style={{
                                            display: 'block',
                                            marginBottom: 4,
                                            fontSize: 14,
                                            fontWeight: 500
                                        }}>Value</label>
                                        <textarea
                                            value={editValue}
                                            onChange={e => setEditValue(e.target.value)}
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
                                        <label
                                            style={{display: 'block', marginBottom: 4, fontSize: 14, fontWeight: 500}}>Authorized
                                            Users</label>
                                        <div style={{
                                            display: 'flex',
                                            flexWrap: 'wrap',
                                            gap: 6,
                                            alignItems: 'center',
                                            minHeight: 40,
                                            padding: '8px',
                                            borderRadius: 5,
                                            border: '1px solid #ccc',
                                            background: '#fff'
                                        }}>
                                            {editSelectedUsers.map(userId => {
                                                const user = editUsers.find(u => u.id === userId);
                                                if (!user) return null;
                                                return (
                                                    <div
                                                        key={user.id}
                                                        style={{
                                                            display: 'flex',
                                                            alignItems: 'center',
                                                            gap: 4,
                                                            padding: '4px 8px',
                                                            borderRadius: 4,
                                                            background: '#f0f9ff',
                                                            border: '1px solid #bae6fd',
                                                            fontSize: 12
                                                        }}
                                                    >
                                                        <span style={{color: '#0369a1', fontSize: 10}}>👤</span>
                                                        <span style={{
                                                            fontWeight: 500,
                                                            color: '#0c4a6e'
                                                        }}>{user.username}</span>
                                                        <button
                                                            type="button"
                                                            onClick={() => handleRemoveUser(user.id)}
                                                            style={{
                                                                marginLeft: 4,
                                                                padding: '2px 4px',
                                                                border: 'none',
                                                                background: 'transparent',
                                                                color: '#dc2626',
                                                                cursor: 'pointer',
                                                                fontSize: 14,
                                                                lineHeight: 1,
                                                                display: 'flex',
                                                                alignItems: 'center',
                                                                justifyContent: 'center'
                                                            }}
                                                            title="Remove user"
                                                        >
                                                            ×
                                                        </button>
                                                    </div>
                                                );
                                            })}
                                            <div style={{position: 'relative', display: 'inline-block'}}
                                                 data-user-selector>
                                                <button
                                                    type="button"
                                                    onClick={() => setShowUserSelector(!showUserSelector)}
                                                    style={{
                                                        padding: '4px 8px',
                                                        borderRadius: 4,
                                                        border: '1px solid #2563eb',
                                                        background: '#fff',
                                                        color: '#2563eb',
                                                        fontSize: 12,
                                                        fontWeight: 500,
                                                        cursor: 'pointer'
                                                    }}
                                                >
                                                    + Add User
                                                </button>
                                                {showUserSelector && (
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
                                                        {getAvailableUsers().length === 0 ? (
                                                            <div style={{fontSize: 12, color: '#666', padding: 8}}>No
                                                                available users</div>
                                                        ) : (
                                                            getAvailableUsers().map(user => (
                                                                <button
                                                                    key={user.id}
                                                                    type="button"
                                                                    onClick={() => handleAddUser(user.id)}
                                                                    style={{
                                                                        width: '100%',
                                                                        display: 'flex',
                                                                        alignItems: 'center',
                                                                        gap: 8,
                                                                        padding: '6px 8px',
                                                                        border: 'none',
                                                                        background: 'transparent',
                                                                        cursor: 'pointer',
                                                                        fontSize: 12,
                                                                        textAlign: 'left'
                                                                    }}
                                                                    onMouseEnter={(e) => {
                                                                        e.currentTarget.style.background = '#f9fafb';
                                                                    }}
                                                                    onMouseLeave={(e) => {
                                                                        e.currentTarget.style.background = 'transparent';
                                                                    }}
                                                                >
                                                                    <span
                                                                        style={{fontWeight: 500}}>{user.username}</span>
                                                                </button>
                                                            ))
                                                        )}
                                                    </div>
                                                )}
                                            </div>
                                        </div>
                                    </div>
                                    <div style={{marginBottom: 12}}>
                                        <label
                                            style={{display: 'block', marginBottom: 4, fontSize: 14, fontWeight: 500}}>Authorized
                                            Roles</label>
                                        <div style={{
                                            display: 'flex',
                                            flexWrap: 'wrap',
                                            gap: 6,
                                            alignItems: 'center',
                                            minHeight: 40,
                                            padding: '8px',
                                            borderRadius: 5,
                                            border: '1px solid #ccc',
                                            background: '#fff'
                                        }}>
                                            {editSelectedRoles.map(roleId => {
                                                const role = editRoles.find(r => r.id === roleId);
                                                if (!role) return null;
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
                                                        <span style={{
                                                            fontWeight: 500,
                                                            color: '#374151'
                                                        }}>{role.name}</span>
                                                        <button
                                                            type="button"
                                                            onClick={() => handleRemoveRole(role.id)}
                                                            style={{
                                                                marginLeft: 4,
                                                                padding: '2px 4px',
                                                                border: 'none',
                                                                background: 'transparent',
                                                                color: '#dc2626',
                                                                cursor: 'pointer',
                                                                fontSize: 14,
                                                                lineHeight: 1,
                                                                display: 'flex',
                                                                alignItems: 'center',
                                                                justifyContent: 'center'
                                                            }}
                                                            title="Remove role"
                                                        >
                                                            ×
                                                        </button>
                                                    </div>
                                                );
                                            })}
                                            <div style={{position: 'relative', display: 'inline-block'}}
                                                 data-role-selector>
                                                <button
                                                    type="button"
                                                    onClick={() => setShowRoleSelector(!showRoleSelector)}
                                                    style={{
                                                        padding: '4px 8px',
                                                        borderRadius: 4,
                                                        border: '1px solid #2563eb',
                                                        background: '#fff',
                                                        color: '#2563eb',
                                                        fontSize: 12,
                                                        fontWeight: 500,
                                                        cursor: 'pointer'
                                                    }}
                                                >
                                                    + Add Role
                                                </button>
                                                {showRoleSelector && (
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
                                                        {getAvailableRoles().length === 0 ? (
                                                            <div style={{fontSize: 12, color: '#666', padding: 8}}>No
                                                                available roles</div>
                                                        ) : (
                                                            getAvailableRoles().map(role => (
                                                                <button
                                                                    key={role.id}
                                                                    type="button"
                                                                    onClick={() => handleAddRole(role.id)}
                                                                    style={{
                                                                        width: '100%',
                                                                        display: 'flex',
                                                                        alignItems: 'center',
                                                                        gap: 8,
                                                                        padding: '6px 8px',
                                                                        border: 'none',
                                                                        background: 'transparent',
                                                                        cursor: 'pointer',
                                                                        fontSize: 12,
                                                                        textAlign: 'left'
                                                                    }}
                                                                    onMouseEnter={(e) => {
                                                                        e.currentTarget.style.background = '#f9fafb';
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
                                                                    <span style={{fontWeight: 500}}>{role.name}</span>
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
                                                            ))
                                                        )}
                                                    </div>
                                                )}
                                            </div>
                                        </div>
                                    </div>
                                    {editError && (
                                        <div style={{color: 'red', marginBottom: 10, fontSize: 14}}>{editError}</div>
                                    )}
                                    <div style={{display: 'flex', gap: 12, marginTop: 18}}>
                                        <button
                                            type="button"
                                            onClick={handleCancelEdit}
                                            disabled={editLoading}
                                            style={{
                                                flex: 1,
                                                padding: 10,
                                                borderRadius: 5,
                                                border: '1px solid #ccc',
                                                background: '#f5f5f5',
                                                cursor: editLoading ? 'not-allowed' : 'pointer'
                                            }}
                                        >
                                            Cancel
                                        </button>
                                        <button
                                            type="submit"
                                            disabled={editLoading}
                                            style={{
                                                flex: 2,
                                                padding: 10,
                                                borderRadius: 5,
                                                border: 'none',
                                                background: '#2563eb',
                                                color: '#fff',
                                                fontWeight: 600,
                                                cursor: editLoading ? 'not-allowed' : 'pointer'
                                            }}
                                        >
                                            {editLoading ? 'Updating...' : 'Update Secret'}
                                        </button>
                                    </div>
                                </form>
                            )}
                        </>
                    ) : null}
                    <button style={{marginTop: 16}} onClick={() => {
                        setSelectedSecret(null);
                        setIsEditing(false);
                        setShowUserSelector(false);
                        setShowRoleSelector(false);
                    }}>Close
                    </button>
                </div>
            )}
            <CreateSecretModal open={showCreateModal} onClose={() => setShowCreateModal(false)}
                               onCreated={fetchSecrets}/>
            <ManageRolesModal open={showManageRolesModal} onClose={() => setShowManageRolesModal(false)}/>
            <ManageUsersModal open={showManageUsersModal} onClose={() => setShowManageUsersModal(false)}/>
        </div>
    );
} 