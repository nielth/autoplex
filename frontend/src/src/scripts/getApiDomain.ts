export function getApiDomain() {
  return import.meta.env.VITE_GO_BACKEND_LOCATION || "";
}
