import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { mockApiClient } from './mock-api-client';

/**
 * Scenario test helpers for common user flows
 */

export interface TestUser {
  id: string;
  email: string;
  name: string;
  organizations: string[];
}

export const defaultTestUser: TestUser = {
  id: 'test-user-123',
  email: 'test@example.com',
  name: 'Test User',
  organizations: ['org-1', 'org-2'],
};

/**
 * Helper to simulate user login
 */
export async function loginUser(user: TestUser = defaultTestUser) {
  const loginButton = screen.getByRole('button', { name: /sign in/i });
  await userEvent.click(loginButton);
  
  await waitFor(() => {
    expect(mockApiClient.auth.login).toHaveBeenCalled();
  });
}

/**
 * Helper to navigate to a workspace
 */
export async function navigateToWorkspace(workspaceId: string) {
  const workspaceLink = screen.getByRole('link', { name: new RegExp(workspaceId) });
  await userEvent.click(workspaceLink);
  
  await waitFor(() => {
    expect(screen.getByTestId('workspace-detail')).toBeInTheDocument();
  });
}

/**
 * Helper to create a new project
 */
export async function createProject(projectData: {
  name: string;
  description?: string;
}) {
  const createButton = screen.getByRole('button', { name: /create project/i });
  await userEvent.click(createButton);
  
  const nameInput = screen.getByLabelText(/project name/i);
  await userEvent.type(nameInput, projectData.name);
  
  if (projectData.description) {
    const descInput = screen.getByLabelText(/description/i);
    await userEvent.type(descInput, projectData.description);
  }
  
  const submitButton = screen.getByRole('button', { name: /create/i });
  await userEvent.click(submitButton);
  
  await waitFor(() => {
    expect(mockApiClient.projects.create).toHaveBeenCalled();
  });
}

/**
 * Helper to deploy an application
 */
export async function deployApplication(appData: {
  name: string;
  type: 'stateless' | 'stateful' | 'cronjob' | 'function';
  image?: string;
}) {
  const deployButton = screen.getByRole('button', { name: /deploy/i });
  await userEvent.click(deployButton);
  
  const nameInput = screen.getByLabelText(/application name/i);
  await userEvent.type(nameInput, appData.name);
  
  const typeSelect = screen.getByLabelText(/application type/i);
  await userEvent.selectOptions(typeSelect, appData.type);
  
  if (appData.image) {
    const imageInput = screen.getByLabelText(/image/i);
    await userEvent.type(imageInput, appData.image);
  }
  
  const submitButton = screen.getByRole('button', { name: /deploy/i });
  await userEvent.click(submitButton);
  
  await waitFor(() => {
    expect(mockApiClient.applications.create).toHaveBeenCalled();
  });
}

/**
 * Helper to trigger a CronJob
 */
export async function triggerCronJob(cronJobId: string) {
  const triggerButton = screen.getByRole('button', { 
    name: new RegExp(`trigger.*${cronJobId}`, 'i') 
  });
  await userEvent.click(triggerButton);
  
  await waitFor(() => {
    expect(mockApiClient.applications.triggerCronJob).toHaveBeenCalled();
  });
}

/**
 * Helper to setup monitoring alerts
 */
export async function setupMonitoringAlert(alertData: {
  name: string;
  metric: string;
  threshold: number;
}) {
  const createAlertButton = screen.getByRole('button', { name: /create alert/i });
  await userEvent.click(createAlertButton);
  
  const nameInput = screen.getByLabelText(/alert name/i);
  await userEvent.type(nameInput, alertData.name);
  
  const metricSelect = screen.getByLabelText(/metric/i);
  await userEvent.selectOptions(metricSelect, alertData.metric);
  
  const thresholdInput = screen.getByLabelText(/threshold/i);
  await userEvent.clear(thresholdInput);
  await userEvent.type(thresholdInput, alertData.threshold.toString());
  
  const submitButton = screen.getByRole('button', { name: /create/i });
  await userEvent.click(submitButton);
  
  // Note: createAlert method doesn't exist in the current monitoring API
  // This would need to be implemented in the monitoring API
  await waitFor(() => {
    expect(screen.getByText(/alert created/i)).toBeInTheDocument();
  });
}