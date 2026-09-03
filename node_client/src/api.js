export class ApiError extends Error {
  constructor(problem, status = 0) {
    const detail =
      problem?.detail ||
      problem?.error ||
      "The request could not be completed.";
    super(detail);
    this.name = "ApiError";
    this.problem = problem;
    this.status = problem?.status || status;
    this.code = problem?.code || "UNKNOWN_ERROR";
    this.invalidParams =
      problem?.invalid_params || problem?.invalidParams || [];
  }
}

export async function apiRequest(path, options = {}) {
  const response = await fetch(path, {
    credentials: "include",
    ...options,
    headers: {
      ...(options.body ? { "Content-Type": "application/json" } : {}),
      ...options.headers,
    },
  });

  if (response.ok) {
    if (response.status === 204) return null;
    return response.json();
  }

  let problem;
  try {
    problem = await response.json();
  } catch {
    problem = {
      detail: `Request failed with status ${response.status}.`,
      status: response.status,
      code: "HTTP_ERROR",
    };
  }
  throw new ApiError(problem, response.status);
}

export function getFieldError(error, fieldName) {
  return error?.invalidParams?.find((param) => param.name === fieldName)
    ?.reason;
}
