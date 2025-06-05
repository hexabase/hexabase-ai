'use client';

import { useState, useEffect } from 'react';
import { Plus, MoreVertical, UserPlus, Shield, Eye, Code, Trash2 } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';
import {
  ProjectMember,
  AddProjectMemberRequest,
  projectMembersApi,
} from '@/lib/api-client';
import { Button } from '@/components/ui/button';
import { LoadingSpinner } from '@/components/ui/loading';

interface ProjectMembersProps {
  organizationId: string;
  projectId: string;
  projectName: string;
}

const roleIcons = {
  admin: Shield,
  developer: Code,
  viewer: Eye,
};

const roleDescriptions = {
  admin: 'Full access to project settings and members',
  developer: 'Can manage namespaces and resources',
  viewer: 'Read-only access to project resources',
};

export function ProjectMembers({ organizationId, projectId, projectName }: ProjectMembersProps) {
  const { toast } = useToast();
  const [members, setMembers] = useState<ProjectMember[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAddMember, setShowAddMember] = useState(false);
  const [addingMember, setAddingMember] = useState(false);
  const [editingMember, setEditingMember] = useState<string | null>(null);
  const [newMember, setNewMember] = useState<AddProjectMemberRequest>({
    user_email: '',
    role: 'developer',
  });

  useEffect(() => {
    loadMembers();
  }, [organizationId, projectId]);

  const loadMembers = async () => {
    try {
      const response = await projectMembersApi.list(organizationId, projectId);
      setMembers(response.members);
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to load project members',
        variant: 'destructive',
      });
      console.error('Error loading members:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleAddMember = async () => {
    if (!newMember.user_email) {
      toast({
        title: 'Error',
        description: 'Please enter a user email',
        variant: 'destructive',
      });
      return;
    }

    setAddingMember(true);
    try {
      const member = await projectMembersApi.add(organizationId, projectId, newMember);
      setMembers([...members, member]);
      setShowAddMember(false);
      setNewMember({ user_email: '', role: 'developer' });
      toast({
        title: 'Success',
        description: 'Member added successfully',
      });
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to add member',
        variant: 'destructive',
      });
      console.error('Error adding member:', error);
    } finally {
      setAddingMember(false);
    }
  };

  const handleUpdateRole = async (memberId: string, newRole: 'admin' | 'developer' | 'viewer') => {
    try {
      const updatedMember = await projectMembersApi.updateRole(organizationId, projectId, memberId, newRole);
      setMembers(members.map(m => m.id === memberId ? updatedMember : m));
      setEditingMember(null);
      toast({
        title: 'Success',
        description: 'Member role updated',
      });
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to update member role',
        variant: 'destructive',
      });
      console.error('Error updating role:', error);
    }
  };

  const handleRemoveMember = async (memberId: string, memberEmail: string) => {
    if (!confirm(`Are you sure you want to remove ${memberEmail} from this project?`)) {
      return;
    }

    try {
      await projectMembersApi.remove(organizationId, projectId, memberId);
      setMembers(members.filter(m => m.id !== memberId));
      toast({
        title: 'Success',
        description: 'Member removed successfully',
      });
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to remove member',
        variant: 'destructive',
      });
      console.error('Error removing member:', error);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h3 className="text-lg font-semibold">Project Members</h3>
          <p className="text-sm text-gray-500">Manage who has access to {projectName}</p>
        </div>
        <Button 
          onClick={() => setShowAddMember(true)}
          className="flex items-center gap-2"
        >
          <UserPlus className="h-4 w-4" />
          Add Member
        </Button>
      </div>

      {/* Add Member Form */}
      {showAddMember && (
        <div className="bg-gray-50 p-4 rounded-lg space-y-4">
          <h4 className="font-medium">Add New Member</h4>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Email Address
              </label>
              <input
                type="email"
                value={newMember.user_email}
                onChange={(e) => setNewMember({ ...newMember, user_email: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
                placeholder="user@example.com"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Role
              </label>
              <select
                value={newMember.role}
                onChange={(e) => setNewMember({ ...newMember, role: e.target.value as any })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md"
              >
                <option value="viewer">Viewer</option>
                <option value="developer">Developer</option>
                <option value="admin">Admin</option>
              </select>
            </div>
          </div>
          <div className="flex gap-2">
            <Button
              onClick={handleAddMember}
              disabled={addingMember}
            >
              {addingMember ? 'Adding...' : 'Add Member'}
            </Button>
            <Button
              variant="secondary"
              onClick={() => {
                setShowAddMember(false);
                setNewMember({ user_email: '', role: 'developer' });
              }}
            >
              Cancel
            </Button>
          </div>
        </div>
      )}

      {/* Members List */}
      <div className="border rounded-lg overflow-hidden">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Member
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Role
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Added
              </th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {members.map((member) => {
              const RoleIcon = roleIcons[member.role];
              return (
                <tr key={member.id}>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div>
                      <div className="text-sm font-medium text-gray-900">
                        {member.user_name}
                      </div>
                      <div className="text-sm text-gray-500">
                        {member.user_email}
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    {editingMember === member.id ? (
                      <select
                        value={member.role}
                        onChange={(e) => handleUpdateRole(member.id, e.target.value as any)}
                        onBlur={() => setEditingMember(null)}
                        className="px-2 py-1 border border-gray-300 rounded-md text-sm"
                        autoFocus
                      >
                        <option value="viewer">Viewer</option>
                        <option value="developer">Developer</option>
                        <option value="admin">Admin</option>
                      </select>
                    ) : (
                      <button
                        onClick={() => setEditingMember(member.id)}
                        className="flex items-center gap-2 text-sm hover:bg-gray-50 px-2 py-1 rounded"
                      >
                        <RoleIcon className="h-4 w-4 text-gray-400" />
                        <span className="capitalize">{member.role}</span>
                      </button>
                    )}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {new Date(member.added_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-right text-sm">
                    <button
                      onClick={() => handleRemoveMember(member.id, member.user_email)}
                      className="text-red-600 hover:text-red-900"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {/* Role Descriptions */}
      <div className="bg-gray-50 p-4 rounded-lg">
        <h4 className="font-medium mb-3">Role Permissions</h4>
        <div className="space-y-2">
          {Object.entries(roleDescriptions).map(([role, description]) => {
            const Icon = roleIcons[role as keyof typeof roleIcons];
            return (
              <div key={role} className="flex items-start gap-2">
                <Icon className="h-4 w-4 text-gray-400 mt-0.5" />
                <div>
                  <span className="font-medium capitalize">{role}:</span>{' '}
                  <span className="text-gray-600">{description}</span>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}