'use client'

import React, {useEffect, useState} from 'react';
import {getSecret, listSecrets} from '../api/secrets';
import {FolderNode, SecretInfoSummary} from '../types';
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
                    <span style={{flex: 1, fontWeight: 500}}>{secret.path.split('/').pop()}</span>
                    <span style={{
                        color: '#888',
                        fontSize: 12,
                        marginLeft: 8
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
            return;
        }
        setSecretInfoLoading(true);
        setSecretInfoError(null);
        setSecretInfo(null);
        setShowValue(false);
        setSecretValue(null);
        setSecretValueError(null);
        getSecret(btoa(selectedSecret.path))
            .then(info => setSecretInfo(info))
            .catch(e => setSecretInfoError(e?.body || 'Failed to fetch secret info'))
            .finally(() => setSecretInfoLoading(false));
    }, [selectedSecret]);

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
                                    <span style={{marginLeft: 8, display: 'flex', alignItems: 'center', gap: 8}}>
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
                        </>
                    ) : null}
                    <button style={{marginTop: 16}} onClick={() => setSelectedSecret(null)}>Close</button>
                </div>
            )}
            <CreateSecretModal open={showCreateModal} onClose={() => setShowCreateModal(false)}
                               onCreated={fetchSecrets}/>
            <ManageRolesModal open={showManageRolesModal} onClose={() => setShowManageRolesModal(false)}/>
            <ManageUsersModal open={showManageUsersModal} onClose={() => setShowManageUsersModal(false)}/>
        </div>
    );
} 