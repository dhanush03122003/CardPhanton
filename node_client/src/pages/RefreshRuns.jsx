import { useEffect, useState } from "react";
import { apiRequest } from "../api";

const formatDate = (value) => {
  if (!value) return "Unknown";
  return new Intl.DateTimeFormat("en-IN", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "Asia/Kolkata",
  }).format(new Date(value));
};

const maskPan = (pan) => {
  const value = String(pan || "");
  return value.length > 4 ? `•••• ${value.slice(-4)}` : value || "Unknown";
};

const statusClass = {
  RUNNING: "bg-amber-100 text-amber-800",
  COMPLETED: "bg-emerald-100 text-emerald-800",
  FAILED: "bg-red-100 text-red-800",
};

function RefreshRuns() {
  const [runs, setRuns] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const loadRuns = async () => {
    setLoading(true);
    setError("");
    try {
      const data = await apiRequest("/api/admin/refresh-runs");
      setRuns(Array.isArray(data?.runs) ? data.runs : []);
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadRuns();
  }, []);

  return (
    <section className='card'>
      <div className='mb-6 flex items-center justify-between border-b border-gray-100 pb-4 dark:border-slate-700'>
        <div>
          <h2 className='font-semibold text-gray-900 dark:text-white'>
            Refresh Runs
          </h2>
          <p className='mt-1 text-sm text-gray-500'>
            Monitor card limit synchronization workflows.
          </p>
        </div>
        <button
          type='button'
          onClick={loadRuns}
          disabled={loading}
          className='btn-secondary'
        >
          {loading ? "Loading..." : "Refresh"}
        </button>
      </div>
      {error && (
        <div className='mb-4 flex items-center justify-between rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700'>
          <span>{error}</span>
          <button type='button' onClick={loadRuns} className='font-semibold'>
            Retry
          </button>
        </div>
      )}
      {loading && runs.length === 0 ? (
        <div className='h-24 animate-pulse rounded-lg bg-gray-50 dark:bg-slate-800' />
      ) : runs.length === 0 ? (
        <p className='py-12 text-center text-sm text-gray-500'>
          No refresh runs recorded yet.
        </p>
      ) : (
        <div className='overflow-x-auto rounded-xl border border-gray-200 dark:border-slate-700'>
          <table className='w-full min-w-[900px] text-left text-sm'>
            <thead>
              <tr className='border-b border-gray-200 bg-gray-50 text-xs uppercase tracking-wide text-gray-500 dark:border-slate-700 dark:bg-slate-800'>
                <th className='px-3 py-3'>Card</th>
                <th className='px-3 py-3'>Status</th>
                <th className='px-3 py-3'>Workflow ID</th>
                <th className='px-3 py-3'>Run ID</th>
                <th className='px-3 py-3'>Started</th>
                <th className='px-3 py-3'>Updated</th>
              </tr>
            </thead>
            <tbody>
              {runs.map((run) => (
                <tr
                  key={run.id}
                  className='border-b border-gray-100 dark:border-slate-800'
                >
                  <td className='px-3 py-4'>
                    <div className='font-mono text-gray-900 dark:text-white'>
                      {maskPan(run.pan)}
                    </div>
                    <div className='mt-1 max-w-48 truncate text-xs text-gray-500'>
                      {run.card_id}
                    </div>
                  </td>
                  <td className='px-3 py-4'>
                    <span
                      className={`rounded-full px-2.5 py-1 text-xs font-semibold ${statusClass[run.status] || "bg-gray-100 text-gray-700"}`}
                    >
                      {run.status}
                    </span>
                  </td>
                  <td className='max-w-48 truncate px-3 py-4 font-mono text-xs text-gray-600'>
                    {run.workflow_id}
                  </td>
                  <td className='max-w-48 truncate px-3 py-4 font-mono text-xs text-gray-600'>
                    {run.run_id}
                  </td>
                  <td className='whitespace-nowrap px-3 py-4 text-gray-700 dark:text-slate-300'>
                    {formatDate(run.created_at)}
                  </td>
                  <td className='whitespace-nowrap px-3 py-4 text-gray-700 dark:text-slate-300'>
                    {formatDate(run.updated_at)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}

export default RefreshRuns;
