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

interface EditScheduleDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  orgId: string
  workspaceId: string
  cronJobId: string
  currentSchedule: string
  onSuccess?: () => void
}

export function EditScheduleDialog({
  open,
  onOpenChange,
  orgId,
  workspaceId,
  cronJobId,
  currentSchedule,
  onSuccess,
}: EditScheduleDialogProps) {
  const { toast } = useToast()
  const [loading, setLoading] = useState(false)
  const [schedule, setSchedule] = useState(currentSchedule)
  const [schedulePreview, setSchedulePreview] = useState('')
  const [cronError, setCronError] = useState('')

  useEffect(() => {
    setSchedule(currentSchedule)
  }, [currentSchedule, open])

  useEffect(() => {
    validateAndPreviewCron(schedule)
  }, [schedule])

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

  const handleUpdate = async () => {
    if (cronError) {
      toast({
        title: 'Error',
        description: cronError,
        variant: 'destructive',
      })
      return
    }

    try {
      setLoading(true)
      
      await apiClient.applicationsApi.updateCronSchedule(orgId, workspaceId, cronJobId, {
        schedule: schedule
      })
      
      toast({
        title: 'Success',
        description: 'Schedule updated successfully',
      })
      
      onSuccess?.()
      onOpenChange(false)
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to update schedule',
        variant: 'destructive',
      })
    } finally {
      setLoading(false)
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
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Edit Schedule</DialogTitle>
          <DialogDescription>
            Update the cron schedule for this job.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <div className="flex items-center gap-2">
              <Label htmlFor="schedule">Schedule (Cron Expression)</Label>
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
                id="schedule"
                name="cron_schedule"
                value={schedule}
                onChange={(e) => setSchedule(e.target.value)}
                disabled={loading}
                className={cronError ? 'border-red-500' : ''}
              />
              <Select
                value={schedule}
                onValueChange={setSchedule}
                disabled={loading}
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

          <div className="text-sm text-muted-foreground">
            <p>Current schedule: <code className="bg-muted px-1 rounded">{currentSchedule}</code></p>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={loading}>
            Cancel
          </Button>
          <Button onClick={handleUpdate} disabled={loading || !!cronError}>
            {loading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Update Schedule
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}