"use client";

import React, { useState, useEffect } from 'react';
import { useAuth } from '@/lib/auth-context';
import { apiClient } from '@/lib/api-client';

interface SecuritySession {
  id: string;
  device: string;
  ip_address: string;
  location?: string;
  last_active: string;
  is_current: boolean;
}

interface SecurityLog {
  id: string;
  event_type: string;
  description: string;
  ip_address: string;
  timestamp: string;
  severity: 'info' | 'warning' | 'error';
}

export function SecuritySettings() {
  const { user, token } = useAuth();
  const [sessions, setSessions] = useState<SecuritySession[]>([]);
  const [securityLogs, setSecurityLogs] = useState<SecurityLog[]>([]);
  const [twoFactorEnabled, setTwoFactorEnabled] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (user && token) {
      fetchSecurityData();
    }
  }, [user, token]);

  const fetchSecurityData = async () => {
    try {
      setLoading(true);
      // Fetch active sessions
      const sessionsResponse = await apiClient.get('/auth/sessions');
      setSessions(sessionsResponse.data.sessions);

      // Fetch security logs
      const logsResponse = await apiClient.get('/auth/security-logs');
      setSecurityLogs(logsResponse.data.logs);

      // Check 2FA status
      const settingsResponse = await apiClient.get('/auth/security-settings');
      setTwoFactorEnabled(settingsResponse.data.two_factor_enabled);
    } catch (error) {
      console.error('Failed to fetch security data:', error);
    } finally {
      setLoading(false);
    }
  };

  const revokeSession = async (sessionId: string) => {
    try {
      await apiClient.delete(`/auth/sessions/${sessionId}`);
      // Refresh sessions list
      await fetchSecurityData();
    } catch (error) {
      console.error('Failed to revoke session:', error);
    }
  };

  const revokeAllSessions = async () => {
    if (confirm('This will log you out of all devices except this one. Continue?')) {
      try {
        await apiClient.post('/auth/sessions/revoke-all');
        await fetchSecurityData();
      } catch (error) {
        console.error('Failed to revoke all sessions:', error);
      }
    }
  };

  const toggleTwoFactor = async () => {
    try {
      if (twoFactorEnabled) {
        // Disable 2FA
        await apiClient.post('/auth/2fa/disable');
        setTwoFactorEnabled(false);
      } else {
        // Enable 2FA - would typically show QR code setup
        const response = await apiClient.post('/auth/2fa/enable');
        // Handle QR code display for 2FA setup
        console.log('2FA setup:', response.data);
        setTwoFactorEnabled(true);
      }
    } catch (error) {
      console.error('Failed to toggle 2FA:', error);
    }
  };

  if (loading) {
    return <div className="p-4">Loading security settings...</div>;
  }

  return (
    <div className="max-w-4xl mx-auto p-6 space-y-8">
      <div>
        <h2 className="text-2xl font-bold mb-6">Security Settings</h2>
        
        {/* Two-Factor Authentication */}
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <h3 className="text-lg font-semibold mb-4">Two-Factor Authentication</h3>
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-600">
                Add an extra layer of security to your account by enabling two-factor authentication.
              </p>
              <p className="text-sm text-gray-500 mt-1">
                Status: {twoFactorEnabled ? (
                  <span className="text-green-600 font-medium">Enabled</span>
                ) : (
                  <span className="text-yellow-600 font-medium">Disabled</span>
                )}
              </p>
            </div>
            <button
              onClick={toggleTwoFactor}
              className={`px-4 py-2 rounded-md font-medium ${
                twoFactorEnabled
                  ? 'bg-red-100 text-red-700 hover:bg-red-200'
                  : 'bg-blue-600 text-white hover:bg-blue-700'
              }`}
            >
              {twoFactorEnabled ? 'Disable 2FA' : 'Enable 2FA'}
            </button>
          </div>
        </div>

        {/* Active Sessions */}
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold">Active Sessions</h3>
            <button
              onClick={revokeAllSessions}
              className="text-sm text-red-600 hover:text-red-700 font-medium"
            >
              Revoke All Other Sessions
            </button>
          </div>
          <div className="space-y-3">
            {sessions.map((session) => (
              <div
                key={session.id}
                className="flex items-center justify-between p-3 border rounded-lg"
              >
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <span className="font-medium">{session.device}</span>
                    {session.is_current && (
                      <span className="text-xs bg-green-100 text-green-700 px-2 py-1 rounded">
                        Current Session
                      </span>
                    )}
                  </div>
                  <p className="text-sm text-gray-500">
                    {session.ip_address} • {session.location || 'Unknown location'}
                  </p>
                  <p className="text-xs text-gray-400">
                    Last active: {new Date(session.last_active).toLocaleString()}
                  </p>
                </div>
                {!session.is_current && (
                  <button
                    onClick={() => revokeSession(session.id)}
                    className="text-red-600 hover:text-red-700 font-medium text-sm"
                  >
                    Revoke
                  </button>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Security Activity Log */}
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold mb-4">Recent Security Activity</h3>
          <div className="space-y-2">
            {securityLogs.slice(0, 10).map((log) => (
              <div
                key={log.id}
                className="flex items-start justify-between p-3 border-b last:border-b-0"
              >
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <span
                      className={`w-2 h-2 rounded-full ${
                        log.severity === 'error'
                          ? 'bg-red-500'
                          : log.severity === 'warning'
                          ? 'bg-yellow-500'
                          : 'bg-green-500'
                      }`}
                    />
                    <span className="font-medium text-sm">{log.event_type}</span>
                  </div>
                  <p className="text-sm text-gray-600 mt-1">{log.description}</p>
                  <p className="text-xs text-gray-400 mt-1">
                    {log.ip_address} • {new Date(log.timestamp).toLocaleString()}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}