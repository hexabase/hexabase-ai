module.exports = {
  format: jest.fn((date, formatString) => {
    // Simple mock that returns a formatted date string
    const d = new Date(date);
    const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
    const month = months[d.getMonth()];
    const day = d.getDate();
    const year = d.getFullYear();
    const hours = String(d.getHours()).padStart(2, '0');
    const minutes = String(d.getMinutes()).padStart(2, '0');
    
    if (formatString === 'MMM d, yyyy HH:mm') {
      return `${month} ${day}, ${year} ${hours}:${minutes}`;
    }
    
    return date.toString();
  })
};