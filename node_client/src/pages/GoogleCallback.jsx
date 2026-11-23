import { useEffect } from "react";

function getBackendOrigin() {
  if (import.meta.env.VITE_API_ORIGIN) {
    return import.meta.env.VITE_API_ORIGIN;
  }
  return window.location.hostname === "localhost"
    ? "http://localhost:8080"
    : window.location.origin;
}

function GoogleCallback() {
  useEffect(() => {
    const callbackURL = new URL(
      "/api/auth/google/callback",
      getBackendOrigin(),
    );
    callbackURL.search = window.location.search;
    window.location.replace(callbackURL.toString());
  }, []);

  return (
    <div className='flex min-h-[calc(100vh-4rem)] items-center justify-center p-6'>
      <p className='text-sm text-gray-500'>Completing Google consent...</p>
    </div>
  );
}

export default GoogleCallback;
