export function getApiDomain() {
  return import.meta.env.VITE_FLASK_LOCATION || "";
}
