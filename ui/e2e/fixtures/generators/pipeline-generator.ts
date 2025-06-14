import { BaseGenerator } from './base-generator';

export interface Pipeline {
  id: string;
  projectId: string;
  name: string;
  description?: string;
  source: {
    type: 'github' | 'gitlab' | 'bitbucket' | 'git';
    repository: string;
    branch?: string;
    path?: string;
    webhook?: {
      id: string;
      events: string[];
      secret?: string;
    };
  };
  trigger: {
    type: 'push' | 'pull_request' | 'tag' | 'schedule' | 'manual';
    branches?: string[];
    tags?: string[];
    schedule?: string;
    filters?: {
      paths?: string[];
      ignorePaths?: string[];
    };
  };
  stages: Array<{
    name: string;
    jobs: Array<{
      name: string;
      image?: string;
      script: string[];
      artifacts?: {
        paths: string[];
        expireIn?: string;
      };
      cache?: {
        key: string;
        paths: string[];
      };
      dependencies?: string[];
      when?: 'always' | 'on_success' | 'on_failure' | 'manual';
      retry?: number;
      timeout?: string;
      parallel?: number;
      environment?: Record<string, string>;
    }>;
    when?: 'always' | 'on_success' | 'on_failure';
  }>;
  variables?: Record<string, string>;
  status: 'pending' | 'running' | 'success' | 'failed' | 'canceled' | 'skipped';
  currentRun?: {
    id: string;
    number: number;
    status: Pipeline['status'];
    startedAt: Date;
    finishedAt?: Date;
    duration?: number;
    commit: {
      sha: string;
      message: string;
      author: string;
      timestamp: Date;
    };
    stages: Array<{
      name: string;
      status: Pipeline['status'];
      startedAt?: Date;
      finishedAt?: Date;
      jobs: Array<{
        name: string;
        status: Pipeline['status'];
        logs?: string;
        exitCode?: number;
      }>;
    }>;
  };
  history?: Array<{
    id: string;
    number: number;
    status: Pipeline['status'];
    startedAt: Date;
    finishedAt: Date;
    duration: number;
    triggeredBy: string;
    commit: string;
  }>;
  statistics?: {
    totalRuns: number;
    successfulRuns: number;
    failedRuns: number;
    averageDuration: number;
    successRate: number;
  };
  createdAt: Date;
  updatedAt: Date;
  createdBy: string;
}

export class PipelineGenerator extends BaseGenerator<Pipeline> {
  private pipelineTemplates = [
    {
      name: 'Node.js CI/CD',
      stages: [
        {
          name: 'build',
          jobs: [{
            name: 'install-and-build',
            image: 'node:18',
            script: ['npm ci', 'npm run build', 'npm run lint'],
          }],
        },
        {
          name: 'test',
          jobs: [{
            name: 'unit-tests',
            image: 'node:18',
            script: ['npm run test:unit'],
          }, {
            name: 'integration-tests',
            image: 'node:18',
            script: ['npm run test:integration'],
          }],
        },
        {
          name: 'deploy',
          jobs: [{
            name: 'deploy-to-k8s',
            image: 'bitnami/kubectl:latest',
            script: ['kubectl apply -f k8s/'],
            when: 'manual',
          }],
        },
      ],
    },
    {
      name: 'Python ML Pipeline',
      stages: [
        {
          name: 'prepare',
          jobs: [{
            name: 'install-deps',
            image: 'python:3.11',
            script: ['pip install -r requirements.txt', 'python -m pytest tests/'],
          }],
        },
        {
          name: 'train',
          jobs: [{
            name: 'train-model',
            image: 'tensorflow/tensorflow:latest',
            script: ['python train.py --epochs 100', 'python evaluate.py'],
          }],
        },
        {
          name: 'deploy',
          jobs: [{
            name: 'deploy-model',
            image: 'google/cloud-sdk:latest',
            script: ['gcloud ai models upload'],
          }],
        },
      ],
    },
    {
      name: 'Docker Build & Push',
      stages: [
        {
          name: 'build',
          jobs: [{
            name: 'docker-build',
            image: 'docker:latest',
            script: [
              'docker build -t $IMAGE_NAME:$CI_COMMIT_SHA .',
              'docker tag $IMAGE_NAME:$CI_COMMIT_SHA $IMAGE_NAME:latest',
            ],
          }],
        },
        {
          name: 'scan',
          jobs: [{
            name: 'security-scan',
            image: 'aquasec/trivy:latest',
            script: ['trivy image $IMAGE_NAME:$CI_COMMIT_SHA'],
          }],
        },
        {
          name: 'push',
          jobs: [{
            name: 'docker-push',
            image: 'docker:latest',
            script: [
              'docker push $IMAGE_NAME:$CI_COMMIT_SHA',
              'docker push $IMAGE_NAME:latest',
            ],
          }],
        },
      ],
    },
  ];
  
  generate(overrides?: Partial<Pipeline>): Pipeline {
    const template = this.faker.helpers.arrayElement(this.pipelineTemplates);
    const name = overrides?.name || `${template.name}-${this.faker.word.adjective()}`;
    
    const pipeline: Pipeline = {
      id: this.generateId('pipe'),
      projectId: overrides?.projectId || this.generateId('proj'),
      name,
      description: overrides?.description || `Automated ${template.name} pipeline`,
      source: {
        type: 'github',
        repository: `${this.faker.internet.userName()}/${this.faker.hacker.noun()}`,
        branch: 'main',
        webhook: {
          id: this.generateId('webhook'),
          events: ['push', 'pull_request'],
          secret: this.faker.string.alphanumeric(32),
        },
      },
      trigger: {
        type: this.faker.helpers.arrayElement(['push', 'pull_request']),
        branches: ['main', 'develop'],
        filters: {
          paths: ['src/**', 'package.json'],
          ignorePaths: ['docs/**', '*.md'],
        },
      },
      stages: this.enhanceStages(template.stages),
      variables: this.generateVariables(template.name),
      status: this.faker.helpers.weighted(
        ['success', 'running', 'failed', 'pending'],
        [0.5, 0.2, 0.2, 0.1]
      ),
      createdAt: this.faker.date.past({ years: 1 }),
      updatedAt: this.faker.date.recent({ days: 7 }),
      createdBy: this.faker.internet.email(),
      ...overrides,
    };
    
    // Generate current run if status is not pending
    if (pipeline.status !== 'pending') {
      pipeline.currentRun = this.generatePipelineRun(pipeline.stages, pipeline.status);
    }
    
    // Generate history
    pipeline.history = this.generateRunHistory();
    pipeline.statistics = this.calculateStatistics(pipeline.history);
    
    return pipeline;
  }
  
  withTraits(traits: string[]): Pipeline {
    const overrides: Partial<Pipeline> = {};
    
    if (traits.includes('failing')) {
      overrides.status = 'failed';
      overrides.currentRun = this.generatePipelineRun([], 'failed');
    }
    
    if (traits.includes('scheduled')) {
      overrides.trigger = {
        type: 'schedule',
        schedule: '0 2 * * *', // Daily at 2 AM
      };
    }
    
    if (traits.includes('complex')) {
      overrides.stages = [
        {
          name: 'prepare',
          jobs: this.generateMultipleJobs(['lint', 'security-check', 'dependency-check']),
        },
        {
          name: 'build',
          jobs: this.generateMultipleJobs(['build-frontend', 'build-backend', 'build-docs']),
        },
        {
          name: 'test',
          jobs: this.generateMultipleJobs(['unit-tests', 'integration-tests', 'e2e-tests', 'performance-tests']),
        },
        {
          name: 'deploy-staging',
          jobs: [{ 
            name: 'deploy-to-staging',
            script: ['./deploy.sh staging'],
            when: 'on_success',
          }],
        },
        {
          name: 'deploy-production',
          jobs: [{
            name: 'deploy-to-production',
            script: ['./deploy.sh production'],
            when: 'manual',
          }],
        },
      ];
    }
    
    if (traits.includes('parallelTests')) {
      const testStage = {
        name: 'test',
        jobs: Array.from({ length: 5 }, (_, i) => ({
          name: `test-suite-${i + 1}`,
          image: 'node:18',
          script: [`npm run test:suite:${i + 1}`],
          parallel: 5,
        })),
      };
      overrides.stages = [
        { name: 'build', jobs: [{ name: 'build', script: ['npm run build'] }] },
        testStage,
      ];
    }
    
    return this.generate(overrides);
  }
  
  private enhanceStages(stages: any[]): Pipeline['stages'] {
    return stages.map(stage => ({
      ...stage,
      jobs: stage.jobs.map(job => ({
        ...job,
        artifacts: job.name.includes('build') ? {
          paths: ['dist/', 'build/'],
          expireIn: '1 week',
        } : undefined,
        cache: job.name.includes('install') || job.name.includes('deps') ? {
          key: '${CI_COMMIT_REF_SLUG}',
          paths: ['node_modules/', '.npm/', 'vendor/'],
        } : undefined,
        retry: job.name.includes('test') ? 2 : undefined,
        timeout: '30m',
        environment: {
          CI: 'true',
          NODE_ENV: stage.name === 'deploy' ? 'production' : 'test',
        },
      })),
    }));
  }
  
  private generateVariables(templateName: string): Record<string, string> {
    const common = {
      CI_REGISTRY: 'registry.example.com',
      IMAGE_NAME: '${CI_REGISTRY}/${CI_PROJECT_PATH}',
      DEPLOY_USER: 'deploy-bot',
    };
    
    const specific: Record<string, Record<string, string>> = {
      'Node.js CI/CD': {
        NODE_VERSION: '18',
        NPM_TOKEN: '${SECRET_NPM_TOKEN}',
        COVERAGE_THRESHOLD: '80',
      },
      'Python ML Pipeline': {
        PYTHON_VERSION: '3.11',
        MODEL_BUCKET: 'gs://ml-models',
        EXPERIMENT_NAME: 'exp-${CI_COMMIT_SHORT_SHA}',
      },
      'Docker Build & Push': {
        DOCKER_DRIVER: 'overlay2',
        DOCKERFILE_PATH: './Dockerfile',
        SCAN_SEVERITY: 'HIGH,CRITICAL',
      },
    };
    
    return { ...common, ...(specific[templateName] || {}) };
  }
  
  private generatePipelineRun(stages: Pipeline['stages'], status: Pipeline['status']): Pipeline['currentRun'] {
    const startedAt = this.faker.date.recent({ days: 1 });
    const isRunning = status === 'running';
    const duration = isRunning ? undefined : this.faker.number.int({ min: 60, max: 1800 });
    
    return {
      id: this.generateId('run'),
      number: this.faker.number.int({ min: 100, max: 999 }),
      status,
      startedAt,
      finishedAt: duration ? new Date(startedAt.getTime() + duration * 1000) : undefined,
      duration,
      commit: {
        sha: this.faker.git.commitSha(),
        message: this.faker.git.commitMessage(),
        author: this.faker.internet.email(),
        timestamp: this.faker.date.recent({ days: 1 }),
      },
      stages: this.generateStageRuns(stages, status),
    };
  }
  
  private generateStageRuns(stages: Pipeline['stages'], pipelineStatus: Pipeline['status']): any[] {
    const stageRuns = [];
    let shouldFail = pipelineStatus === 'failed';
    let hasFailedStage = false;
    
    for (let i = 0; i < stages.length; i++) {
      const stage = stages[i];
      let stageStatus: Pipeline['status'] = 'success';
      
      if (shouldFail && !hasFailedStage && i === stages.length - 1) {
        // Make last stage fail if pipeline failed
        stageStatus = 'failed';
        hasFailedStage = true;
      } else if (pipelineStatus === 'running' && i === Math.floor(stages.length / 2)) {
        stageStatus = 'running';
      } else if (pipelineStatus === 'running' && i > Math.floor(stages.length / 2)) {
        stageStatus = 'pending';
      }
      
      const startedAt = stageStatus !== 'pending' ? 
        this.faker.date.recent({ days: 1 }) : undefined;
      const duration = stageStatus === 'success' || stageStatus === 'failed' ?
        this.faker.number.int({ min: 30, max: 600 }) : undefined;
      
      stageRuns.push({
        name: stage.name,
        status: stageStatus,
        startedAt,
        finishedAt: startedAt && duration ? 
          new Date(startedAt.getTime() + duration * 1000) : undefined,
        jobs: stage.jobs.map(job => ({
          name: job.name,
          status: stageStatus,
          logs: this.generateJobLogs(job.name, stageStatus),
          exitCode: stageStatus === 'success' ? 0 : 
            stageStatus === 'failed' ? 1 : undefined,
        })),
      });
    }
    
    return stageRuns;
  }
  
  private generateJobLogs(jobName: string, status: string): string {
    const logs = [];
    
    // Starting logs
    logs.push(`[${new Date().toISOString()}] Starting job: ${jobName}`);
    logs.push('Preparing environment...');
    logs.push('Pulling docker image...');
    
    // Job-specific logs
    if (jobName.includes('build')) {
      logs.push('> npm run build');
      logs.push('Building application...');
      logs.push('Webpack: Compiling...');
      logs.push('Webpack: Build completed in 45.2s');
    } else if (jobName.includes('test')) {
      logs.push('> npm run test');
      logs.push('Running test suite...');
      logs.push('  ✓ Component tests (42 passed)');
      logs.push('  ✓ API tests (28 passed)');
      logs.push('  ✓ Integration tests (15 passed)');
    } else if (jobName.includes('deploy')) {
      logs.push('Deploying to Kubernetes...');
      logs.push('kubectl apply -f deployment.yaml');
      logs.push('deployment.apps/app configured');
      logs.push('service/app-service configured');
    }
    
    // Status-specific ending
    if (status === 'failed') {
      logs.push('ERROR: ' + this.faker.helpers.arrayElement([
        'Tests failed: 3 failing tests',
        'Build failed: Module not found',
        'Deploy failed: Unauthorized',
        'Timeout: Job exceeded maximum duration',
      ]));
      logs.push('Job failed with exit code 1');
    } else if (status === 'success') {
      logs.push('Job completed successfully');
    }
    
    return logs.join('\n');
  }
  
  private generateRunHistory(): Pipeline['history'] {
    const historyCount = this.faker.number.int({ min: 5, max: 20 });
    const history = [];
    
    for (let i = 0; i < historyCount; i++) {
      const status = this.faker.helpers.weighted(
        ['success', 'failed', 'canceled'],
        [0.7, 0.2, 0.1]
      );
      
      const startedAt = this.faker.date.past({ years: 0.1 });
      const duration = this.faker.number.int({ min: 60, max: 1800 });
      
      history.push({
        id: this.generateId('run'),
        number: historyCount - i,
        status,
        startedAt,
        finishedAt: new Date(startedAt.getTime() + duration * 1000),
        duration,
        triggeredBy: this.faker.helpers.arrayElement([
          this.faker.internet.email(),
          'webhook',
          'schedule',
          'api',
        ]),
        commit: this.faker.git.commitSha().substring(0, 7),
      });
    }
    
    return history;
  }
  
  private calculateStatistics(history: Pipeline['history']): Pipeline['statistics'] {
    if (!history || history.length === 0) {
      return {
        totalRuns: 0,
        successfulRuns: 0,
        failedRuns: 0,
        averageDuration: 0,
        successRate: 0,
      };
    }
    
    const successful = history.filter(h => h.status === 'success').length;
    const failed = history.filter(h => h.status === 'failed').length;
    const totalDuration = history.reduce((sum, h) => sum + h.duration, 0);
    
    return {
      totalRuns: history.length,
      successfulRuns: successful,
      failedRuns: failed,
      averageDuration: Math.round(totalDuration / history.length),
      successRate: parseFloat(((successful / history.length) * 100).toFixed(1)),
    };
  }
  
  private generateMultipleJobs(names: string[]): Pipeline['stages'][0]['jobs'] {
    return names.map(name => ({
      name,
      image: 'node:18',
      script: [`npm run ${name}`],
      when: 'on_success' as const,
    }));
  }
}