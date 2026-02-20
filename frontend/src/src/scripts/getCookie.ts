export function getCookie(name: string) {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) {
    const cookie = parts.pop();
    if (cookie) {
      return cookie.split(";").shift();
    }
  }
  return "";
}
