import { authHandlers } from './auth';
import { organizationHandlers } from './organizations';
import { workspaceHandlers } from './workspaces';
import { projectHandlers } from './projects';
import { applicationHandlers } from './applications';
import { functionHandlers } from './functions';

export const handlers = [
  ...authHandlers,
  ...organizationHandlers,
  ...workspaceHandlers,
  ...projectHandlers,
  ...applicationHandlers,
  ...functionHandlers,
];