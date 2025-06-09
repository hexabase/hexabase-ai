'use client'

import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Loader2, Info, CalendarClock } from 'lucide-react'
import { useToast } from '@/hooks/use-toast'
import { apiClient } from '@/lib/api-client'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { Alert, AlertDescription } from '@/components/ui/alert'

interface Template {
  id: string
  name: string
  source_image?: string
  config?: any
}

interface CreateCronJobDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  orgId: string
  workspaceId: string
  projectId: string
  onSuccess?: () => void
}

export function CreateCronJobDialog({
  open,
  onOpenChange,
  orgId,
  workspaceId,
  projectId,
  onSuccess,
}: CreateCronJobDialogProps) {
  const { toast } = useToast()
  const [loading, setLoading] = useState(false)
  const [useTemplate, setUseTemplate] = useState(false)
  const [templates, setTemplates] = useState<Template[]>([])
  const [enableBackupIntegration, setEnableBackupIntegration] = useState(false)
  const [backupPolicies, setBackupPolicies] = useState<any[]>([])
  const [formData, setFormData] = useState({
    name: '',
    cron_schedule: '0 2 * * *',
    template_app_id: '',
    source_type: 'image',
    source_image: '',
    command: '',
    args: '',
    backup_policy_id: '',
  })
  const [schedulePreview, setSchedulePreview] = useState('Daily at 2:00 AM')
  const [cronError, setCronError] = useState('')

  useEffect(() => {
    if (open) {
      if (useTemplate) {
        fetchTemplates()
      }
      if (enableBackupIntegration) {
        fetchBackupPolicies()
      }
    }
  }, [open, useTemplate, enableBackupIntegration])

  useEffect(() => {
    validateAndPreviewCron(formData.cron_schedule)
  }, [formData.cron_schedule])

  const fetchTemplates = async () => {
    try {
      const response = await apiClient.applicationsApi.list(orgId, workspaceId, null, {
        type: 'stateless',
        is_template: true
      })
      setTemplates(response.data.applications)
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to fetch templates',
        variant: 'destructive',
      })
    }
  }

  const fetchBackupPolicies = async () => {
    try {
      const response = await apiClient.backupApi.listBackupPolicies(orgId, workspaceId)
      setBackupPolicies(response.data.policies || [])
    } catch (error) {
      console.error('Failed to fetch backup policies:', error)
    }
  }

  const validateAndPreviewCron = (expression: string) => {
    setCronError('')
    
    // Basic validation
    const parts = expression.trim().split(' ')
    if (parts.length !== 5) {
      setCronError('Invalid cron expression. Must have 5 fields.')
      setSchedulePreview('')
      return
    }

    // Common patterns
    const patterns: { [key: string]: string } = {
      '* * * * *': 'Every minute',
      '*/5 * * * *': 'Every 5 minutes',
      '*/10 * * * *': 'Every 10 minutes',
      '*/15 * * * *': 'Every 15 minutes',
      '*/30 * * * *': 'Every 30 minutes',
      '0 * * * *': 'Every hour',
      '0 */2 * * *': 'Every 2 hours',
      '0 */3 * * *': 'Every 3 hours',
      '0 */6 * * *': 'Every 6 hours',
      '0 */12 * * *': 'Every 12 hours',
      '0 0 * * *': 'Daily at midnight',
      '0 2 * * *': 'Daily at 2:00 AM',
      '0 0 * * 0': 'Weekly on Sunday',
      '0 0 * * 1': 'Weekly on Monday',
      '0 0 1 * *': 'Monthly on the 1st',
      '0 0 15 * *': 'Monthly on the 15th',
    }
    
    setSchedulePreview(patterns[expression] || 'Custom schedule')
  }

  const handleCreate = async () => {
    if (!formData.name.trim()) {
      toast({
        title: 'Error',
        description: 'CronJob name is required',
        variant: 'destructive',
      })
      return
    }

    if (cronError) {
      toast({
        title: 'Error',
        description: cronError,
        variant: 'destructive',
      })
      return
    }

    if (!useTemplate && !formData.source_image.trim()) {
      toast({
        title: 'Error',
        description: 'Container image is required',
        variant: 'destructive',
      })
      return
    }

    try {
      setLoading(true)
      
      const command = formData.command.trim() ? formData.command.trim().split(' ') : undefined
      const args = formData.args.trim() ? formData.args.trim().split(' ') : undefined
      
      const payload = {
        name: formData.name,
        type: 'cronjob',
        cron_schedule: formData.cron_schedule,
        cron_command: command,
        cron_args: args,
        template_app_id: useTemplate ? formData.template_app_id : undefined,
        source_type: useTemplate ? undefined : formData.source_type,
        source_image: useTemplate ? undefined : formData.source_image,
        status: 'active',
      }

      await apiClient.applicationsApi.create(orgId, workspaceId, projectId, payload)
      
      toast({
        title: 'Success',
        description: 'CronJob created successfully',
      })
      
      onSuccess?.()
      handleClose()
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to create CronJob',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
    }
  }

  const handleClose = () => {
    setFormData({
      name: '',
      cron_schedule: '0 2 * * *',
      template_app_id: '',
      source_type: 'image',
      source_image: '',
      command: '',
      args: '',
      backup_policy_id: '',
    })
    setUseTemplate(false)
    setEnableBackupIntegration(false)
    setCronError('')
    onOpenChange(false)
  }

  const handleBackupPolicySelect = (policyId: string) => {
    const policy = backupPolicies.find(p => p.id === policyId)
    if (policy) {
      setFormData({
        ...formData,
        backup_policy_id: policyId,
        cron_schedule: policy.schedule,
      })
    }
  }

  const cronExamples = [
    { label: 'Every minute', value: '* * * * *' },
    { label: 'Every 5 minutes', value: '*/5 * * * *' },
    { label: 'Every hour', value: '0 * * * *' },
    { label: 'Daily at 2 AM', value: '0 2 * * *' },
    { label: 'Every 6 hours', value: '0 */6 * * *' },
    { label: 'Weekly on Sunday', value: '0 0 * * 0' },
    { label: 'Monthly on 1st', value: '0 0 1 * *' },
  ]

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[625px]">
        <DialogHeader>
          <DialogTitle>Create CronJob</DialogTitle>
          <DialogDescription>
            Create a new scheduled job that runs on a recurring schedule.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="name">CronJob Name</Label>
            <Input
              id="name"
              name="name"
              placeholder="daily-backup"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              disabled={loading}
            />
          </div>

          <div className="grid gap-2">
            <div className="flex items-center gap-2">
              <Label htmlFor="cron_schedule">Schedule (Cron Expression)</Label>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger>
                    <Info className="h-4 w-4 text-muted-foreground" />
                  </TooltipTrigger>
                  <TooltipContent className="max-w-xs">
                    <p>Use standard cron syntax: minute hour day month weekday</p>
                    <p className="mt-1 text-xs">Fields: 0-59 0-23 1-31 1-12 0-7</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
            <div className="flex gap-2">
              <Input
                id="cron_schedule"
                name="cron_schedule"
                value={formData.cron_schedule}
                onChange={(e) => setFormData({ ...formData, cron_schedule: e.target.value })}
                disabled={loading || enableBackupIntegration}
                className={cronError ? 'border-red-500' : ''}
              />
              <Select
                value={formData.cron_schedule}
                onValueChange={(value) => setFormData({ ...formData, cron_schedule: value })}
                disabled={loading || enableBackupIntegration}
              >
                <SelectTrigger className="w-[180px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {cronExamples.map((ex) => (
                    <SelectItem key={ex.value} value={ex.value}>
                      {ex.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            {schedulePreview && !cronError && (
              <p className="text-sm text-muted-foreground flex items-center gap-1">
                <CalendarClock className="h-3 w-3" />
                {schedulePreview}
              </p>
            )}
            {cronError && (
              <p className="text-sm text-red-500">{cronError}</p>
            )}
          </div>

          <div className="space-y-2">
            <div className="flex items-center space-x-2">
              <Checkbox
                id="use_template"
                name="use_template"
                checked={useTemplate}
                onCheckedChange={(checked) => setUseTemplate(checked as boolean)}
                disabled={loading}
              />
              <label
                htmlFor="use_template"
                className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
              >
                Create from template
              </label>
            </div>

            {useTemplate && (
              <div className="grid gap-2 ml-6">
                <Label htmlFor="template_app_id">Template</Label>
                <Select
                  name="template_app_id"
                  value={formData.template_app_id}
                  onValueChange={(value) => setFormData({ ...formData, template_app_id: value })}
                  disabled={loading}
                >
                  <SelectTrigger id="template_app_id">
                    <SelectValue placeholder="Select a template" />
                  </SelectTrigger>
                  <SelectContent>
                    {templates.map((template) => (
                      <SelectItem key={template.id} value={template.id}>
                        <div>
                          <p>{template.name}</p>
                          {template.source_image && (
                            <p className="text-xs text-muted-foreground">{template.source_image}</p>
                          )}
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}
          </div>

          {!useTemplate && (
            <div className="grid gap-2">
              <Label htmlFor="source_image">Container Image</Label>
              <Input
                id="source_image"
                name="source_image"
                placeholder="busybox:latest"
                value={formData.source_image}
                onChange={(e) => setFormData({ ...formData, source_image: e.target.value })}
                disabled={loading}
              />
            </div>
          )}

          <div className="grid gap-2">
            <Label htmlFor="command">Command (Optional)</Label>
            <Input
              id="command"
              name="command"
              placeholder="/bin/sh -c"
              value={formData.command}
              onChange={(e) => setFormData({ ...formData, command: e.target.value })}
              disabled={loading}
            />
            <p className="text-xs text-muted-foreground">
              Override the default container command
            </p>
          </div>

          <div className="grid gap-2">
            <Label htmlFor="args">Arguments (Optional)</Label>
            <Textarea
              id="args"
              name="args"
              placeholder="echo 'Hello from CronJob'"
              value={formData.args}
              onChange={(e) => setFormData({ ...formData, args: e.target.value })}
              disabled={loading}
              rows={2}
            />
            <p className="text-xs text-muted-foreground">
              Space-separated arguments for the command
            </p>
          </div>

          <div className="space-y-2">
            <div className="flex items-center space-x-2">
              <Checkbox
                id="enable_backup_integration"
                name="enable_backup_integration"
                checked={enableBackupIntegration}
                onCheckedChange={(checked) => setEnableBackupIntegration(checked as boolean)}
                disabled={loading}
              />
              <label
                htmlFor="enable_backup_integration"
                className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
              >
                Link with backup policy
              </label>
            </div>

            {enableBackupIntegration && backupPolicies.length > 0 && (
              <div className="grid gap-2 ml-6">
                <Label htmlFor="backup_policy_id">Backup Policy</Label>
                <Select
                  name="backup_policy_id"
                  value={formData.backup_policy_id}
                  onValueChange={handleBackupPolicySelect}
                  disabled={loading}
                >
                  <SelectTrigger id="backup_policy_id">
                    <SelectValue placeholder="Select a backup policy" />
                  </SelectTrigger>
                  <SelectContent>
                    {backupPolicies.map((policy) => (
                      <SelectItem key={policy.id} value={policy.id}>
                        <div>
                          <p>{policy.application_name}</p>
                          <p className="text-xs text-muted-foreground">
                            {policy.schedule} • {policy.retention_days} days retention
                          </p>
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                {formData.backup_policy_id && (
                  <Alert>
                    <Info className="h-4 w-4" />
                    <AlertDescription>
                      Schedule synced with backup policy
                    </AlertDescription>
                  </Alert>
                )}
              </div>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleClose} disabled={loading}>
            Cancel
          </Button>
          <Button 
            onClick={handleCreate} 
            disabled={loading || !!cronError || (!useTemplate && !formData.source_image)}
            data-testid="create-cronjob-button"
          >
            {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Create CronJob
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}