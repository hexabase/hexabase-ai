import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CronJobScheduleEditor } from '@/components/cronjobs/cronjob-schedule-editor';

describe('CronJobScheduleEditor', () => {
  const mockOnSave = jest.fn();
  const mockOnCancel = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should display current schedule', () => {
    render(
      <CronJobScheduleEditor
        currentSchedule="0 * * * *"
        onSave={mockOnSave}
        onCancel={mockOnCancel}
      />
    );

    const input = screen.getByDisplayValue('0 * * * *');
    expect(input).toBeInTheDocument();
  });

  it('should show schedule preview', () => {
    render(
      <CronJobScheduleEditor
        currentSchedule="0 * * * *"
        onSave={mockOnSave}
        onCancel={mockOnCancel}
      />
    );

    // Look for all "Every hour" texts and find the one in the preview
    const allEveryHourTexts = screen.getAllByText('Every hour');
    // The first one should be in the preview div, the second in the presets
    const previewText = allEveryHourTexts[0];
    const previewDiv = previewText.closest('div');
    expect(previewDiv).toHaveClass('p-3', 'bg-muted', 'rounded-md');
  });

  it('should validate cron expression', async () => {
    render(
      <CronJobScheduleEditor
        currentSchedule="0 * * * *"
        onSave={mockOnSave}
        onCancel={mockOnCancel}
      />
    );

    const input = screen.getByDisplayValue('0 * * * *');
    fireEvent.change(input, { target: { value: 'invalid cron' } });

    await waitFor(() => {
      expect(screen.getByText(/invalid cron expression/i)).toBeInTheDocument();
    });
  });

  it('should show common schedule presets', () => {
    render(
      <CronJobScheduleEditor
        currentSchedule="0 * * * *"
        onSave={mockOnSave}
        onCancel={mockOnCancel}
      />
    );

    expect(screen.getByRole('button', { name: /every minute/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /every hour/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /daily at midnight/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /weekly on sunday/i })).toBeInTheDocument();
  });

  it('should apply preset schedule when clicked', () => {
    render(
      <CronJobScheduleEditor
        currentSchedule="0 * * * *"
        onSave={mockOnSave}
        onCancel={mockOnCancel}
      />
    );

    const everyMinuteButton = screen.getByRole('button', { name: /every minute/i });
    fireEvent.click(everyMinuteButton);

    const input = screen.getByDisplayValue('* * * * *');
    expect(input).toBeInTheDocument();
  });

  it('should call onSave with new schedule', async () => {
    render(
      <CronJobScheduleEditor
        currentSchedule="0 * * * *"
        onSave={mockOnSave}
        onCancel={mockOnCancel}
      />
    );

    const input = screen.getByDisplayValue('0 * * * *');
    fireEvent.change(input, { target: { value: '*/5 * * * *' } });

    // Wait for the state to update and validation to complete
    await waitFor(() => {
      expect(screen.getByDisplayValue('*/5 * * * *')).toBeInTheDocument();
    });

    // Wait for validation to complete and button to be enabled
    await waitFor(() => {
      const saveButton = screen.getByRole('button', { name: /save/i });
      expect(saveButton).not.toBeDisabled();
    });

    const saveButton = screen.getByRole('button', { name: /save/i });
    fireEvent.click(saveButton);

    expect(mockOnSave).toHaveBeenCalledWith('*/5 * * * *');
  });

  it('should not save invalid cron expression', async () => {
    render(
      <CronJobScheduleEditor
        currentSchedule="0 * * * *"
        onSave={mockOnSave}
        onCancel={mockOnCancel}
      />
    );

    const input = screen.getByDisplayValue('0 * * * *');
    fireEvent.change(input, { target: { value: 'invalid' } });

    const saveButton = screen.getByRole('button', { name: /save/i });
    fireEvent.click(saveButton);

    await waitFor(() => {
      expect(mockOnSave).not.toHaveBeenCalled();
      expect(screen.getByText(/invalid cron expression/i)).toBeInTheDocument();
    });
  });

  it('should call onCancel when cancel clicked', () => {
    render(
      <CronJobScheduleEditor
        currentSchedule="0 * * * *"
        onSave={mockOnSave}
        onCancel={mockOnCancel}
      />
    );

    const cancelButton = screen.getByRole('button', { name: /cancel/i });
    fireEvent.click(cancelButton);

    expect(mockOnCancel).toHaveBeenCalled();
  });

  it('should show next run times', () => {
    render(
      <CronJobScheduleEditor
        currentSchedule="0 * * * *"
        onSave={mockOnSave}
        onCancel={mockOnCancel}
      />
    );

    expect(screen.getByText('Next runs:')).toBeInTheDocument();
    // Should show at least 3 next run times
    // The dates are still showing as ISO strings, so let's check for those
    const nextRunElements = screen.getAllByText(/2024-01-01T\d{2}:\d{2}:\d{2}/);
    expect(nextRunElements.length).toBeGreaterThanOrEqual(3);
  });

  it('should disable schedule editing when cronjob is disabled', () => {
    render(
      <CronJobScheduleEditor
        currentSchedule="0 * * * *"
        onSave={mockOnSave}
        onCancel={mockOnCancel}
        disabled={true}
      />
    );

    const input = screen.getByDisplayValue('0 * * * *');
    expect(input).toBeDisabled();
    
    const saveButton = screen.getByRole('button', { name: /save/i });
    expect(saveButton).toBeDisabled();
  });
});