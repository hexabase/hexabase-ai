'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Clock, AlertCircle } from 'lucide-react';
import { parseExpression } from 'cron-parser';
import { format } from 'date-fns';

interface CronJobScheduleEditorProps {
  currentSchedule: string;
  onSave: (schedule: string) => void;
  onCancel: () => void;
  disabled?: boolean;
}

const SCHEDULE_PRESETS = [
  { label: 'Every minute', value: '* * * * *' },
  { label: 'Every hour', value: '0 * * * *' },
  { label: 'Daily at midnight', value: '0 0 * * *' },
  { label: 'Weekly on Sunday', value: '0 0 * * 0' },
  { label: 'Every month', value: '0 0 1 * *' },
  { label: 'Every 5 minutes', value: '*/5 * * * *' },
  { label: 'Every 30 minutes', value: '*/30 * * * *' },
  { label: 'Daily at 2 AM', value: '0 2 * * *' },
];

export function CronJobScheduleEditor({
  currentSchedule,
  onSave,
  onCancel,
  disabled = false,
}: CronJobScheduleEditorProps) {
  const [schedule, setSchedule] = useState(currentSchedule);
  const [error, setError] = useState<string | null>(null);
  const [preview, setPreview] = useState<string | null>(null);
  const [nextRuns, setNextRuns] = useState<Date[]>([]);

  const validateAndPreview = (cronExpression: string) => {
    try {
      const interval = parseExpression(cronExpression);
      
      // Get human-readable description
      const description = getHumanReadableSchedule(cronExpression);
      setPreview(description);
      
      // Get next 5 run times
      const runs: Date[] = [];
      let current = interval.next();
      for (let i = 0; i < 5; i++) {
        runs.push(current.toDate());
        current = interval.next();
      }
      setNextRuns(runs);
      setError(null);
      return true;
    } catch (err) {
      setError('Invalid cron expression');
      setPreview(null);
      setNextRuns([]);
      return false;
    }
  };

  const getHumanReadableSchedule = (cronExpression: string) => {
    const preset = SCHEDULE_PRESETS.find(p => p.value === cronExpression);
    if (preset) return preset.label;

    const parts = cronExpression.split(' ');
    if (parts.length !== 5) return 'Invalid expression';

    const [minute, hour, dayOfMonth, month, dayOfWeek] = parts;

    if (minute === '*' && hour === '*' && dayOfMonth === '*' && month === '*' && dayOfWeek === '*') {
      return 'Every minute';
    }
    if (minute === '0' && hour === '*' && dayOfMonth === '*' && month === '*' && dayOfWeek === '*') {
      return 'Every hour';
    }
    if (hour === '0' && minute === '0' && dayOfMonth === '*' && month === '*' && dayOfWeek === '*') {
      return 'Every day at midnight';
    }

    return `Custom schedule: ${cronExpression}`;
  };

  useEffect(() => {
    validateAndPreview(schedule);
  }, [schedule]);

  const handleSave = () => {
    if (validateAndPreview(schedule)) {
      onSave(schedule);
    }
  };

  const applyPreset = (presetValue: string) => {
    setSchedule(presetValue);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Clock className="h-5 w-5" />
          Edit Schedule
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div>
          <Label htmlFor="schedule">Cron Expression</Label>
          <Input
            id="schedule"
            value={schedule}
            onChange={(e) => setSchedule(e.target.value)}
            placeholder="* * * * *"
            disabled={disabled}
          />
          {error && (
            <div className="flex items-center gap-1 mt-1 text-sm text-destructive">
              <AlertCircle className="h-3 w-3" />
              {error}
            </div>
          )}
        </div>

        {preview && (
          <div className="p-3 bg-muted rounded-md">
            <p className="text-sm font-medium">{preview}</p>
          </div>
        )}

        <div>
          <p className="text-sm font-medium mb-2">Quick Presets</p>
          <div className="flex flex-wrap gap-2">
            {SCHEDULE_PRESETS.map((preset) => (
              <Button
                key={preset.value}
                size="sm"
                variant="outline"
                onClick={() => applyPreset(preset.value)}
                disabled={disabled}
              >
                {preset.label}
              </Button>
            ))}
          </div>
        </div>

        {nextRuns.length > 0 && (
          <div>
            <p className="text-sm font-medium mb-2">Next runs:</p>
            <div className="space-y-1">
              {nextRuns.slice(0, 3).map((run, index) => (
                <div key={index} className="text-sm text-muted-foreground">
                  {format(run, 'MMM d, yyyy HH:mm')}
                </div>
              ))}
            </div>
          </div>
        )}

        <div className="flex justify-end gap-2 pt-4">
          <Button variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={!!error || disabled}>
            Save
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}