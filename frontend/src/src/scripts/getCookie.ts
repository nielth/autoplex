export function getCookie(name: string): string | undefined {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) {
    const cookiePart = parts.pop();
    return cookiePart ? cookiePart.split(";").shift() : undefined;
  }
  return undefined;
}

