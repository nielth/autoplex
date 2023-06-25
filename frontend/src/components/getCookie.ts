export function getCookie(authToken: string) {
  const value = `; ${document.cookie}`;
  console.log("incoming cookie:");
  console.log(value);
  const parts = value.split(`; ${authToken}=`);
  if (parts.length === 2) return parts.pop()?.split(";").shift();
}
