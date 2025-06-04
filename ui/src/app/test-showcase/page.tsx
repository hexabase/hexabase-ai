'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Activity, CreditCard, Package, Users, CheckCircle } from 'lucide-react';

export default function TestShowcasePage() {
  const [activeSection, setActiveSection] = useState<'overview' | 'workspaces' | 'projects' | 'billing' | 'monitoring'>('overview');

  return (
    <div data-testid="showcase-page" className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-7xl mx-auto space-y-8">
        {/* Header */}
        <div className="text-center">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">Hexabase KaaS Platform</h1>
          <p className="text-xl text-gray-600">Complete Frontend Implementation Showcase</p>
          <div className="mt-6 flex justify-center space-x-3">
            <Button 
              variant={activeSection === 'overview' ? 'default' : 'outline'}
              onClick={() => setActiveSection('overview')}
              data-testid="overview-button"
            >
              Overview
            </Button>
            <Button 
              variant={activeSection === 'workspaces' ? 'default' : 'outline'}
              onClick={() => setActiveSection('workspaces')}
              data-testid="workspaces-button"
            >
              Workspaces
            </Button>
            <Button 
              variant={activeSection === 'projects' ? 'default' : 'outline'}
              onClick={() => setActiveSection('projects')}
              data-testid="projects-button"
            >
              Projects
            </Button>
            <Button 
              variant={activeSection === 'billing' ? 'default' : 'outline'}
              onClick={() => setActiveSection('billing')}
              data-testid="billing-button"
            >
              Billing
            </Button>
            <Button 
              variant={activeSection === 'monitoring' ? 'default' : 'outline'}
              onClick={() => setActiveSection('monitoring')}
              data-testid="monitoring-button"
            >
              Monitoring
            </Button>
          </div>
        </div>

        {/* Overview Section */}
        {activeSection === 'overview' && (
          <div data-testid="overview-section" className="space-y-6">
            <h2 className="text-2xl font-bold text-center">Platform Overview</h2>
            
            {/* Stats Grid */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Organizations</CardTitle>
                  <Users className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">5</div>
                  <p className="text-xs text-muted-foreground">Active organizations</p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Workspaces</CardTitle>
                  <Package className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">12</div>
                  <p className="text-xs text-muted-foreground">vClusters deployed</p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Projects</CardTitle>
                  <Activity className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">28</div>
                  <p className="text-xs text-muted-foreground">Active projects</p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium">Monthly Cost</CardTitle>
                  <CreditCard className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">$2,450</div>
                  <p className="text-xs text-muted-foreground">Current billing</p>
                </CardContent>
              </Card>
            </div>

            {/* Features Completed */}
            <div className="bg-white rounded-lg border p-6">
              <h3 className="text-lg font-semibold mb-4">Completed Features</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="flex items-center space-x-2">
                  <CheckCircle className="w-5 h-5 text-green-500" />
                  <span>OAuth Authentication (Google & GitHub)</span>
                </div>
                <div className="flex items-center space-x-2">
                  <CheckCircle className="w-5 h-5 text-green-500" />
                  <span>Organization Management</span>
                </div>
                <div className="flex items-center space-x-2">
                  <CheckCircle className="w-5 h-5 text-green-500" />
                  <span>Workspace (vCluster) Management</span>
                </div>
                <div className="flex items-center space-x-2">
                  <CheckCircle className="w-5 h-5 text-green-500" />
                  <span>Project & Namespace Management</span>
                </div>
                <div className="flex items-center space-x-2">
                  <CheckCircle className="w-5 h-5 text-green-500" />
                  <span>Billing & Subscription System</span>
                </div>
                <div className="flex items-center space-x-2">
                  <CheckCircle className="w-5 h-5 text-green-500" />
                  <span>Real-time Monitoring Dashboard</span>
                </div>
                <div className="flex items-center space-x-2">
                  <CheckCircle className="w-5 h-5 text-green-500" />
                  <span>Alerts & Incident Management</span>
                </div>
                <div className="flex items-center space-x-2">
                  <CheckCircle className="w-5 h-5 text-green-500" />
                  <span>Advanced Logs Viewer</span>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Workspaces Section */}
        {activeSection === 'workspaces' && (
          <div data-testid="workspaces-section" className="space-y-6">
            <h2 className="text-2xl font-bold text-center">Workspace Management</h2>
            
            <div className="bg-white rounded-lg border p-6">
              <h3 className="font-semibold mb-4">Active Workspaces</h3>
              <div className="space-y-3">
                <div className="border rounded-lg p-4 flex items-center justify-between">
                  <div>
                    <h4 className="font-medium">Production Workspace</h4>
                    <p className="text-sm text-gray-600">vCluster: prod-cluster-v1</p>
                  </div>
                  <span className="px-3 py-1 bg-green-100 text-green-800 rounded-full text-sm">Active</span>
                </div>
                <div className="border rounded-lg p-4 flex items-center justify-between">
                  <div>
                    <h4 className="font-medium">Staging Workspace</h4>
                    <p className="text-sm text-gray-600">vCluster: staging-cluster-v1</p>
                  </div>
                  <span className="px-3 py-1 bg-green-100 text-green-800 rounded-full text-sm">Active</span>
                </div>
                <div className="border rounded-lg p-4 flex items-center justify-between">
                  <div>
                    <h4 className="font-medium">Development Workspace</h4>
                    <p className="text-sm text-gray-600">vCluster: dev-cluster-v1</p>
                  </div>
                  <span className="px-3 py-1 bg-yellow-100 text-yellow-800 rounded-full text-sm">Provisioning</span>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Projects Section */}
        {activeSection === 'projects' && (
          <div data-testid="projects-section" className="space-y-6">
            <h2 className="text-2xl font-bold text-center">Project Management</h2>
            
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="bg-white rounded-lg border p-6">
                <h4 className="font-semibold mb-2">Frontend Application</h4>
                <p className="text-sm text-gray-600 mb-4">React-based frontend with microservices</p>
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span>Namespaces</span>
                    <span className="font-medium">3</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span>Pods</span>
                    <span className="font-medium">12</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span>CPU Usage</span>
                    <span className="font-medium">65%</span>
                  </div>
                </div>
              </div>
              
              <div className="bg-white rounded-lg border p-6">
                <h4 className="font-semibold mb-2">Backend Services</h4>
                <p className="text-sm text-gray-600 mb-4">Microservices API architecture</p>
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span>Namespaces</span>
                    <span className="font-medium">5</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span>Pods</span>
                    <span className="font-medium">23</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span>CPU Usage</span>
                    <span className="font-medium">78%</span>
                  </div>
                </div>
              </div>
              
              <div className="bg-white rounded-lg border p-6">
                <h4 className="font-semibold mb-2">Data Pipeline</h4>
                <p className="text-sm text-gray-600 mb-4">ETL and data processing services</p>
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span>Namespaces</span>
                    <span className="font-medium">2</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span>Pods</span>
                    <span className="font-medium">8</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span>CPU Usage</span>
                    <span className="font-medium">45%</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Billing Section */}
        {activeSection === 'billing' && (
          <div data-testid="billing-section" className="space-y-6">
            <h2 className="text-2xl font-bold text-center">Billing & Subscription</h2>
            
            <div className="bg-white rounded-lg border p-6">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
                <div>
                  <p className="text-sm text-gray-600">Current Plan</p>
                  <p className="text-2xl font-bold text-blue-600">Enterprise</p>
                  <p className="text-lg">$299/month</p>
                </div>
                <div>
                  <p className="text-sm text-gray-600">Next Billing Date</p>
                  <p className="text-lg font-medium">January 1, 2025</p>
                </div>
                <div>
                  <p className="text-sm text-gray-600">Payment Method</p>
                  <p className="text-lg font-medium">•••• 4242</p>
                </div>
              </div>
              
              <div className="border-t pt-6">
                <h4 className="font-medium mb-4">Resource Usage</h4>
                <div className="space-y-3">
                  <div>
                    <div className="flex justify-between text-sm mb-1">
                      <span>Workspaces</span>
                      <span>12 / 50</span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div className="bg-blue-600 h-2 rounded-full" style={{ width: '24%' }}></div>
                    </div>
                  </div>
                  <div>
                    <div className="flex justify-between text-sm mb-1">
                      <span>Storage</span>
                      <span>450GB / 2TB</span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div className="bg-green-600 h-2 rounded-full" style={{ width: '22.5%' }}></div>
                    </div>
                  </div>
                  <div>
                    <div className="flex justify-between text-sm mb-1">
                      <span>Bandwidth</span>
                      <span>1.2TB / 10TB</span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div className="bg-purple-600 h-2 rounded-full" style={{ width: '12%' }}></div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Monitoring Section */}
        {activeSection === 'monitoring' && (
          <div data-testid="monitoring-section" className="space-y-6">
            <h2 className="text-2xl font-bold text-center">Monitoring & Observability</h2>
            
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <div className="bg-white rounded-lg border p-4">
                <div className="flex items-center justify-between mb-2">
                  <h4 className="font-medium">Cluster Health</h4>
                  <CheckCircle className="w-5 h-5 text-green-500" />
                </div>
                <p className="text-2xl font-bold">Healthy</p>
                <p className="text-sm text-gray-600">All systems operational</p>
              </div>
              
              <div className="bg-white rounded-lg border p-4">
                <h4 className="font-medium mb-2">CPU Usage</h4>
                <p className="text-2xl font-bold">42.5%</p>
                <div className="w-full bg-gray-200 rounded-full h-2 mt-2">
                  <div className="bg-blue-600 h-2 rounded-full" style={{ width: '42.5%' }}></div>
                </div>
              </div>
              
              <div className="bg-white rounded-lg border p-4">
                <h4 className="font-medium mb-2">Memory Usage</h4>
                <p className="text-2xl font-bold">68.3%</p>
                <div className="w-full bg-gray-200 rounded-full h-2 mt-2">
                  <div className="bg-purple-600 h-2 rounded-full" style={{ width: '68.3%' }}></div>
                </div>
              </div>
              
              <div className="bg-white rounded-lg border p-4">
                <h4 className="font-medium mb-2">Active Alerts</h4>
                <p className="text-2xl font-bold text-yellow-600">2</p>
                <p className="text-sm text-gray-600">Requires attention</p>
              </div>
            </div>
            
            <div className="bg-white rounded-lg border p-6">
              <h4 className="font-medium mb-4">Recent Alerts</h4>
              <div className="space-y-3">
                <div className="flex items-center justify-between p-3 bg-yellow-50 rounded-lg">
                  <div className="flex items-center space-x-3">
                    <div className="w-2 h-2 bg-yellow-500 rounded-full"></div>
                    <div>
                      <p className="font-medium">High Memory Usage</p>
                      <p className="text-sm text-gray-600">Production workspace at 85% memory</p>
                    </div>
                  </div>
                  <span className="text-sm text-gray-500">15 min ago</span>
                </div>
                <div className="flex items-center justify-between p-3 bg-blue-50 rounded-lg">
                  <div className="flex items-center space-x-3">
                    <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
                    <div>
                      <p className="font-medium">Deployment Successful</p>
                      <p className="text-sm text-gray-600">Frontend app v2.3.1 deployed</p>
                    </div>
                  </div>
                  <span className="text-sm text-gray-500">1 hour ago</span>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Success Message */}
        <div className="bg-green-50 border border-green-200 rounded-lg p-6 text-center">
          <CheckCircle className="w-12 h-12 text-green-500 mx-auto mb-3" />
          <h3 className="text-lg font-semibold text-green-800 mb-2">All Features Implemented Successfully!</h3>
          <p className="text-green-700">
            The Hexabase KaaS platform frontend is fully functional with all core features completed.
          </p>
        </div>
      </div>
    </div>
  );
}