module.exports = {
  parseExpression: jest.fn((expression) => {
    // Basic validation
    const parts = expression.split(' ');
    if (parts.length !== 5) {
      throw new Error('Invalid cron expression');
    }
    
    // Return a mock interval object
    const dates = [
      new Date('2024-01-01T00:00:00Z'),
      new Date('2024-01-01T00:05:00Z'),
      new Date('2024-01-01T00:10:00Z'),
      new Date('2024-01-01T00:15:00Z'),
      new Date('2024-01-01T00:20:00Z'),
    ];
    
    let index = 0;
    
    return {
      next: () => ({
        toDate: () => {
          if (index < dates.length) {
            return dates[index++];
          }
          return dates[0];
        }
      })
    };
  })
};