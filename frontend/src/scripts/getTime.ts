export function getTime(time: string) {
    console.log(time);
    const unixTimeZero = Date.parse(time);
    const date = new Date(unixTimeZero);
    return date.toDateString().split(" ").slice(1).join(" ");
  }
  
  export function formatDateTime(isoString: string) {
    // Parse the ISO string into a Date object
    const date = new Date(isoString);
  
    // Extract each component
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, "0"); // Months are 0-indexed
    const day = String(date.getDate()).padStart(2, "0");
    const hour = String(date.getHours()).padStart(2, "0");
    const minute = String(date.getMinutes()).padStart(2, "0");
    const second = String(date.getSeconds()).padStart(2, "0");
  
    // Format and return the result
    return `${year}-${month}-${day} ${hour}:${minute}:${second}`;
  }
  
  export function formatUnixTimestamp(unixTimestamp: number) {
    // Convert Unix timestamp to milliseconds and create a Date object
    const date = new Date(unixTimestamp * 1000);
  
    // Extract each component
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, "0"); // Months are 0-indexed
    const day = String(date.getDate()).padStart(2, "0");
    const hour = String(date.getHours()).padStart(2, "0");
    const minute = String(date.getMinutes()).padStart(2, "0");
    const second = String(date.getSeconds()).padStart(2, "0");
  
    // Format and return the result
    return `${year}-${month}-${day} ${hour}:${minute}:${second}`;
  }
  
  export function timeSince(timestamp: number) {
    // Get the current timestamp in milliseconds
    const now = Date.now();
  
    // Convert the given timestamp from seconds to milliseconds
    const then = timestamp * 1000;
  
    // Calculate the difference in milliseconds
    const diff = now - then;
  
    // Convert the difference to hours
    const hours = diff / (1000 * 60 * 60);
  
    if (hours < 24) {
      // If less than 24 hours, return the hours
      return Math.floor(hours) + " hours ago";
    } else if (hours < 72) {
      // Otherwise, convert the difference to days and return
      const days = hours / 24;
      return Math.floor(days) + " days ago";
    } else {
      return formatUnixTimestamp(timestamp)
    }
  }