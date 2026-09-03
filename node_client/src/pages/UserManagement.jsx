import { useEffect, useState } from "react";
import { ApiError } from "../api";

const getDeviceInfo = (auth, aaguidMap) => {
  const attachmentType = auth.attachmentType || auth.attachment_type;
  const transports = auth.transports || [];
  const aaguid = (auth.aaguid || "").toLowerCase();
  const brand = aaguidMap[aaguid];
  if (brand)
    return {
      name: brand.name,
      logo: brand.icon_light || brand.icon_dark,
      description:
        attachmentType === "platform" ? "Platform passkey" : "Security key",
    };
  if (
    transports.includes("internal") &&
    /Mac|iPhone|iPad/i.test(navigator.userAgent)
  )
    return {
      name: "Apple iCloud Keychain",
      logo: null,
      description: "Face ID / Touch ID",
    };
  return {
    name:
      attachmentType === "platform" ? "Built-in Biometrics" : "Security Key",
    logo: null,
    description: "Standard WebAuthn device",
  };
};

const formatDate = (value) => {
  if (!value) return "Unknown";
  return new Intl.DateTimeFormat("en-IN", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "Asia/Kolkata",
  }).format(new Date(value));
};

function UserManagement() {
  const [adminUsers, setAdminUsers] = useState([]);
  const [users, setUsers] = useState([]);
  const [selectedUser, setSelectedUser] = useState(null);
  const [details, setDetails] = useState(null);
  const [loading, setLoading] = useState(true);
  const [detailsLoading, setDetailsLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState(false);
  const [pageError, setPageError] = useState("");
  const [modalError, setModalError] = useState("");
  const [aaguidMap, setAaguidMap] = useState({});
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(0);
  const [expandedLogId, setExpandedLogId] = useState(null);
  const [userLogsExpanded, setUserLogsExpanded] = useState(true);
  const [adminLogsExpanded, setAdminLogsExpanded] = useState(true);

  const userAuditLogs = details
    ? [...details.user_audit_logs].sort(
        (left, right) => new Date(right.login_time) - new Date(left.login_time),
      )
    : [];
  const adminAuditLogs = details
    ? [...details.admin_audit_logs].sort(
        (left, right) => new Date(right.created_at) - new Date(left.created_at),
      )
    : [];

  const getActionBadgeClass = (actionType = "") => {
    const action = actionType.toUpperCase();
    if (action.includes("DELETE") || action.includes("REJECT")) {
      return "bg-red-100 text-red-700";
    }
    if (action.includes("APPROV") || action.includes("CREATE")) {
      return "bg-emerald-100 text-emerald-700";
    }
    if (action.includes("LOGIN")) return "bg-blue-100 text-blue-700";
    if (action.includes("SUSPEND")) return "bg-amber-100 text-amber-700";
    return "bg-gray-100 text-gray-700";
  };

  const renderAuditLog = (log, isAdminAction) => {
    const timelineId = `${isAdminAction ? "admin" : "user"}-${log.id}`;
    const expanded = expandedLogId === timelineId;
    const timestamp = isAdminAction ? log.created_at : log.login_time;
    const detailFields = isAdminAction
      ? {
          admin_user_id: log.admin_user_id,
          target_user_id: log.target_user_id,
          action_type: log.action_type,
          details: log.details,
          created_at: log.created_at,
        }
      : {
          user_id: log.user_id,
          credential_id: log.credential_id,
          action_type: log.action_type,
          ip_address: log.ip_address,
          location: log.location,
          user_agent: log.user_agent,
          login_time: log.login_time,
        };

    return (
      <div
        key={timelineId}
        className={`overflow-hidden rounded-lg border border-gray-200 border-l-4 ${isAdminAction ? "border-l-violet-400" : "border-l-blue-400"} bg-white shadow-sm transition-colors hover:bg-gray-50`}
      >
        <button
          type='button'
          onClick={() => setExpandedLogId(expanded ? null : timelineId)}
          className='flex w-full items-center gap-3 px-4 py-3 text-left'
        >
          <span
            className={`rounded-full px-2 py-1 text-[10px] font-bold uppercase tracking-wide ${getActionBadgeClass(log.action_type)}`}
          >
            {log.action_type}
          </span>
          <span className='text-xs font-medium text-gray-500'>
            {formatDate(timestamp)}
          </span>
          <span className='min-w-0 flex-1 truncate text-xs text-gray-500'>
            {isAdminAction
              ? log.details || "Administrative action"
              : log.ip_address || log.location || "User activity"}
          </span>
          <span
            className={`text-gray-400 transition-transform ${expanded ? "rotate-180" : ""}`}
            aria-hidden='true'
          >
            ⌄
          </span>
        </button>
        <div
          className={`grid transition-[grid-template-rows] duration-200 ${expanded ? "grid-rows-[1fr]" : "grid-rows-[0fr]"}`}
        >
          <div className='min-h-0 overflow-hidden'>
            <pre className='border-t border-gray-100 bg-gray-50 px-4 py-3 text-xs leading-5 text-gray-600'>
              {JSON.stringify(detailFields, null, 2)}
            </pre>
          </div>
        </div>
      </div>
    );
  };

  const renderAuditSection = (
    title,
    description,
    logs,
    expanded,
    setExpanded,
    isAdminAction,
  ) => (
    <section>
      <button
        type='button'
        onClick={() => setExpanded((value) => !value)}
        className='mb-3 flex w-full items-center justify-between rounded-lg border border-gray-200 bg-gray-50 px-4 py-3 text-left transition-colors hover:bg-gray-100'
        aria-expanded={expanded}
      >
        <div>
          <h4 className='font-semibold text-gray-900'>{title}</h4>
          <p className='mt-1 text-xs text-gray-500'>{description}</p>
        </div>
        <div className='flex items-center gap-3'>
          <span className='rounded-full bg-white px-2 py-1 text-xs font-semibold text-gray-600'>
            {logs.length}
          </span>
          <span
            className={`text-gray-400 transition-transform ${expanded ? "rotate-180" : ""}`}
            aria-hidden='true'
          >
            ⌄
          </span>
        </div>
      </button>
      {expanded &&
        (logs.length === 0 ? (
          <p className='rounded-lg border border-dashed border-gray-200 py-8 text-center text-sm text-gray-500'>
            No logs recorded.
          </p>
        ) : (
          <div className='max-h-96 space-y-2 overflow-y-auto pr-2 [scrollbar-color:#cbd5e1_transparent] [scrollbar-width:thin]'>
            {logs.map((log) => renderAuditLog(log, isAdminAction))}
          </div>
        ))}
    </section>
  );

  const loadUsers = async (requestedPage = page, requestedSearch = search) => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        page: String(requestedPage),
        page_size: String(pageSize),
      });
      if (requestedSearch.trim()) params.set("search", requestedSearch.trim());
      const response = await fetch(`/api/admin/users?${params.toString()}`, {
        credentials: "include",
      });
      const data = await response.json();
      if (!response.ok) throw new ApiError(data, response.status);
      const responseUsers = Array.isArray(data)
        ? data
        : Array.isArray(data.users)
          ? data.users
          : [];
      const responseTotal = Array.isArray(data)
        ? data.length
        : Number.isFinite(data.total)
          ? data.total
          : responseUsers.length;
      const responsePage = Array.isArray(data)
        ? requestedPage
        : data.page || requestedPage;
      const responsePageSize = Array.isArray(data)
        ? pageSize
        : data.page_size || pageSize;
      setUsers(responseUsers);
      setPage(responsePage);
      setTotal(responseTotal);
      setTotalPages(
        Array.isArray(data)
          ? responseTotal > 0
            ? Math.ceil(responseTotal / responsePageSize)
            : 0
          : data.total_pages || 0,
      );
    } catch (requestError) {
      setPageError(requestError.message);
    } finally {
      setLoading(false);
    }
  };

  const loadAdminUsers = async () => {
    try {
      const response = await fetch("/api/admin/admins", {
        credentials: "include",
      });
      const data = await response.json();
      if (!response.ok) throw new ApiError(data, response.status);
      setAdminUsers(Array.isArray(data) ? data : []);
    } catch (requestError) {
      setPageError(requestError.message);
    }
  };

  useEffect(() => {
    loadUsers();
    loadAdminUsers();
    fetch(
      "https://raw.githubusercontent.com/passkeydeveloper/passkey-authenticator-aaguids/main/aaguid.json",
    )
      .then((response) => (response.ok ? response.json() : {}))
      .then(setAaguidMap)
      .catch((requestError) =>
        console.error("Error fetching AAGUID map:", requestError),
      );
  }, []);

  const handleSearchChange = (event) => {
    const value = event.target.value;
    setSearch(value);
    setPage(1);
    loadUsers(1, value);
  };

  const changePage = (nextPage) => {
    if (nextPage < 1 || nextPage > totalPages) return;
    setPage(nextPage);
    loadUsers(nextPage, search);
  };

  const openUser = async (user) => {
    setSelectedUser(user);
    setDetails(null);
    setDetailsLoading(true);
    setModalError("");
    try {
      const response = await fetch(`/api/admin/users/${user.id}/details`, {
        credentials: "include",
      });
      const data = await response.json();
      if (!response.ok) throw new ApiError(data, response.status);
      data.authenticators = Array.isArray(data.authenticators)
        ? data.authenticators
        : [];
      data.user_audit_logs = Array.isArray(data.user_audit_logs)
        ? data.user_audit_logs
        : [];
      data.admin_audit_logs = Array.isArray(data.admin_audit_logs)
        ? data.admin_audit_logs
        : [];
      data.cards = Array.isArray(data.cards) ? data.cards : [];
      setDetails(data);
      if (data.user) setSelectedUser(data.user);
    } catch (requestError) {
      setModalError(requestError.message);
    } finally {
      setDetailsLoading(false);
    }
  };

  const runAction = async (path, method, successMessage) => {
    setActionLoading(true);
    setModalError("");
    try {
      const response = await fetch(path, { method, credentials: "include" });
      const data = await response.json();
      if (!response.ok) throw new ApiError(data, response.status);
      alert(successMessage);
      await loadUsers();
      if (selectedUser) await openUser(selectedUser);
    } catch (requestError) {
      setModalError(requestError.message);
    } finally {
      setActionLoading(false);
    }
  };

  const deleteUser = async () => {
    if (
      !selectedUser ||
      !window.confirm(`Delete ${selectedUser.username}? This cannot be undone.`)
    )
      return;
    setActionLoading(true);
    try {
      const response = await fetch(`/api/admin/users/${selectedUser.id}`, {
        method: "DELETE",
        credentials: "include",
      });
      const data = await response.json();
      if (!response.ok) throw new ApiError(data, response.status);
      setUsers((items) => items.filter((item) => item.id !== selectedUser.id));
      setSelectedUser(null);
      setDetails(null);
    } catch (requestError) {
      setModalError(requestError.message);
    } finally {
      setActionLoading(false);
    }
  };

  const deleteAuthenticator = (authId) =>
    runAction(
      `/api/admin/users/${selectedUser.id}/authenticators/${authId}`,
      "DELETE",
      "Authenticator deleted.",
    );

  const normalUsers = users;

  const renderUserRow = (item) => (
    <tr
      key={item.id}
      onClick={() => openUser(item)}
      className='cursor-pointer hover:bg-gray-50'
    >
      <td className='px-3 py-4 text-xs text-gray-500'>{item.id}</td>
      <td className='px-3 py-4 text-sm font-medium text-gray-900'>
        {item.username}
      </td>
      <td className='px-3 py-4'>
        <span
          className={`rounded-full px-2 py-1 text-xs font-semibold ${item.status === "ENABLED" ? "bg-emerald-100 text-emerald-700" : "bg-amber-100 text-amber-700"}`}
        >
          {item.status}
        </span>
      </td>
    </tr>
  );

  const renderUserTable = (title, items, emptyMessage, compact = false) => (
    <div
      className={`overflow-x-auto rounded-lg border border-gray-200 ${compact ? "border-0" : ""}`}
    >
      <div className='border-b border-gray-200 bg-gray-50 px-4 py-3'>
        <h3 className='text-sm font-semibold text-gray-800'>
          {title} ({items.length})
        </h3>
      </div>
      <table className='w-full text-left'>
        <thead>
          <tr className='border-b border-gray-200 text-xs uppercase tracking-wide text-gray-500'>
            <th className='px-3 py-3'>User ID</th>
            <th className='px-3 py-3'>Username</th>
            <th className='px-3 py-3'>Status</th>
          </tr>
        </thead>
        <tbody className='divide-y divide-gray-100'>
          {items.map(renderUserRow)}
        </tbody>
      </table>
      {items.length === 0 && (
        <p className='py-8 text-center text-sm text-gray-500'>{emptyMessage}</p>
      )}
    </div>
  );

  return (
    <section className='card'>
      <div className='flex items-center justify-between mb-6 pb-4 border-b border-gray-100'>
        <div>
          <h2 className='text-base font-semibold text-gray-900'>
            User management
          </h2>
          <p className='text-sm text-gray-500 mt-1'>
            Review accounts, passkeys, and administrative history.
          </p>
        </div>
        <button
          onClick={() => {
            loadAdminUsers();
            loadUsers();
          }}
          className='btn-secondary'
          disabled={loading}
        >
          Refresh
        </button>
      </div>
      {pageError && !selectedUser && (
        <p className='mb-4 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700'>
          {pageError}
        </p>
      )}
      {loading ? (
        <div className='h-24 animate-pulse rounded-lg bg-gray-50' />
      ) : (
        <div className='space-y-6'>
          <div className='rounded-xl border border-gray-200 bg-white p-4 shadow-sm'>
            {renderUserTable(
              "Admins",
              adminUsers,
              "No admin users found.",
              true,
            )}
          </div>
          <div className='space-y-4 rounded-xl border border-gray-200 bg-white p-4 shadow-sm'>
            <div className='flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
              <input
                type='search'
                value={search}
                onChange={handleSearchChange}
                placeholder='Search normal users by username or ID'
                aria-label='Search normal users'
                className='input-field sm:max-w-sm'
              />
              {search.trim() && (
                <span className='text-xs text-gray-500'>
                  {total} result{total === 1 ? "" : "s"} found
                </span>
              )}
            </div>
            {renderUserTable(
              "Normal Users",
              normalUsers,
              search ? "No users match your search." : "No normal users found.",
              true,
            )}
            {totalPages > 0 && (
              <div className='flex items-center justify-between border-t border-gray-100 pt-4'>
                <button
                  type='button'
                  onClick={() => changePage(page - 1)}
                  disabled={loading || page <= 1}
                  className='btn-secondary disabled:cursor-not-allowed disabled:opacity-50'
                >
                  Previous
                </button>
                <span className='text-sm text-gray-600'>
                  Page {page} of {totalPages}
                </span>
                <button
                  type='button'
                  onClick={() => changePage(page + 1)}
                  disabled={loading || page >= totalPages}
                  className='btn-secondary disabled:cursor-not-allowed disabled:opacity-50'
                >
                  Next
                </button>
              </div>
            )}
          </div>
        </div>
      )}

      {selectedUser && (
        <div
          className='fixed inset-0 z-50 flex items-center justify-center bg-slate-900/40 p-4'
          onClick={() => setSelectedUser(null)}
        >
          <div
            className='max-h-[90vh] w-full max-w-4xl overflow-y-auto rounded-xl bg-white p-6 shadow-xl'
            onClick={(event) => event.stopPropagation()}
          >
            <div className='flex items-start justify-between border-b border-gray-100 pb-4'>
              <div>
                <h3 className='text-lg font-semibold text-gray-900'>
                  {selectedUser.username}
                </h3>
                <p className='text-xs text-gray-500'>{selectedUser.id}</p>
              </div>
              <button
                onClick={() => setSelectedUser(null)}
                className='text-sm text-gray-500 hover:text-gray-900'
              >
                Close
              </button>
            </div>
            {modalError && (
              <p className='mt-4 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700'>
                {modalError}
              </p>
            )}
            {detailsLoading ? (
              <div className='h-32 animate-pulse bg-gray-50 mt-6' />
            ) : (
              details && (
                <div className='space-y-7 pt-6'>
                  <section>
                    <div className='mb-3 flex items-center justify-between'>
                      <div>
                        <h4 className='font-semibold text-gray-900'>
                          Associated Cards
                        </h4>
                        <p className='mt-1 text-xs text-gray-500'>
                          Cards linked to this user account.
                        </p>
                      </div>
                      <span className='rounded-full bg-gray-100 px-2 py-1 text-xs font-semibold text-gray-600'>
                        {details.cards.length}
                      </span>
                    </div>
                    {details.cards.length === 0 ? (
                      <p className='rounded-lg border border-dashed border-gray-200 py-8 text-center text-sm text-gray-500'>
                        No cards associated with this user.
                      </p>
                    ) : (
                      <div className='overflow-x-auto rounded-lg border border-gray-200'>
                        <table className='w-full text-left text-sm'>
                          <thead>
                            <tr className='border-b border-gray-200 bg-gray-50 text-xs uppercase tracking-wide text-gray-500'>
                              <th className='px-3 py-3'>Card</th>
                              <th className='px-3 py-3'>Product</th>
                              <th className='px-3 py-3'>Cardholder</th>
                              <th className='px-3 py-3'>Bank</th>
                              <th className='px-3 py-3'>Expires</th>
                              <th className='px-3 py-3'>Type</th>
                              <th className='px-3 py-3'>CVV</th>
                              <th className='px-3 py-3'>Linked Phone</th>
                            </tr>
                          </thead>
                          <tbody className='divide-y divide-gray-100'>
                            {details.cards.map((card) => (
                              <tr key={card.id}>
                                <td className='px-3 py-3 font-medium text-gray-900'>
                                  ••••{" "}
                                  {String(card.pan || "").slice(-4) ||
                                    "Unknown"}
                                  <div className='text-xs text-gray-500'>
                                    {card.card_brand}
                                  </div>
                                </td>
                                <td className='px-3 py-3 font-medium text-gray-900'>
                                  {card.product_name || "Standard card"}
                                </td>
                                <td className='px-3 py-3 text-gray-700'>
                                  {card.cardholder_name}
                                </td>
                                <td className='px-3 py-3 text-gray-700'>
                                  {card.bank_name}
                                </td>
                                <td className='px-3 py-3 text-gray-700'>
                                  {card.exp_month}/{card.exp_year}
                                </td>
                                <td className='px-3 py-3 text-gray-700'>
                                  {card.payment_method_type}
                                </td>
                                <td className='px-3 py-3 font-mono text-gray-700'>
                                  {card.cvv || "Unknown"}
                                </td>
                                <td className='px-3 py-3 font-mono text-gray-700'>
                                  {card.linked_phone_number || "Unknown"}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    )}
                  </section>
                  <section>
                    <div className='mb-3 flex items-center justify-between'>
                      <h4 className='font-semibold text-gray-900'>
                        Authenticators
                      </h4>
                    </div>
                    <div className='overflow-x-auto'>
                      <table className='w-full text-left text-sm'>
                        <thead>
                          <tr className='border-b text-xs text-gray-500'>
                            <th className='px-2 py-2'>Nickname</th>
                            <th className='px-2 py-2'>Created</th>
                            <th className='px-2 py-2 text-right'>Action</th>
                          </tr>
                        </thead>
                        <tbody className='divide-y'>
                          {details.authenticators.map((auth) => {
                            const info = getDeviceInfo(auth, aaguidMap);
                            return (
                              <tr key={auth.id}>
                                <td className='px-2 py-3'>
                                  <div className='flex items-center gap-3'>
                                    <div className='flex h-9 w-9 items-center justify-center rounded-lg border border-gray-100 bg-gray-50'>
                                      {info.logo ? (
                                        <img
                                          src={info.logo}
                                          alt=''
                                          className='h-5 w-5 object-contain'
                                        />
                                      ) : (
                                        <span>
                                          {info.name.includes("Apple")
                                            ? "🍏"
                                            : info.name.includes("Biometric")
                                              ? "👤"
                                              : "🔑"}
                                        </span>
                                      )}
                                    </div>
                                    <div>
                                      <div>
                                        {auth.nickname || "Unnamed passkey"}
                                      </div>
                                      <div className='text-xs text-gray-500'>
                                        {info.name} · {info.description}
                                      </div>
                                    </div>
                                  </div>
                                </td>
                                <td className='px-2 py-3 text-gray-500'>
                                  {formatDate(auth.created_at)}
                                </td>
                                <td className='px-2 py-3 text-right'>
                                  <button
                                    onClick={() => deleteAuthenticator(auth.id)}
                                    disabled={actionLoading}
                                    className='text-xs text-red-600'
                                  >
                                    Delete
                                  </button>
                                </td>
                              </tr>
                            );
                          })}
                        </tbody>
                      </table>
                    </div>
                  </section>
                  {renderAuditSection(
                    "User Activity Logs",
                    "Authentication and user activity for this account.",
                    userAuditLogs,
                    userLogsExpanded,
                    setUserLogsExpanded,
                    false,
                  )}
                  {renderAuditSection(
                    "Admin Actions",
                    "Administrative actions performed on this account.",
                    adminAuditLogs,
                    adminLogsExpanded,
                    setAdminLogsExpanded,
                    true,
                  )}
                  <section className='border-t border-red-100 pt-5'>
                    <h4 className='font-semibold text-red-700'>
                      Account actions
                    </h4>
                    <div className='mt-3 flex gap-2'>
                      {selectedUser.status === "DISABLED" && (
                        <button
                          onClick={() =>
                            runAction(
                              `/api/admin/users/${selectedUser.id}/approve`,
                              "PUT",
                              "User approved.",
                            )
                          }
                          disabled={actionLoading}
                          className='btn-secondary text-emerald-700'
                        >
                          Approve User
                        </button>
                      )}
                      {selectedUser.status === "ENABLED" && (
                        <button
                          onClick={() =>
                            runAction(
                              `/api/admin/users/${selectedUser.id}/suspend`,
                              "PUT",
                              "User suspended.",
                            )
                          }
                          disabled={actionLoading}
                          className='btn-secondary text-amber-700'
                        >
                          Suspend User
                        </button>
                      )}
                      <button
                        onClick={deleteUser}
                        disabled={actionLoading}
                        className='btn-secondary text-red-600'
                      >
                        Delete User
                      </button>
                    </div>
                  </section>
                </div>
              )
            )}
          </div>
        </div>
      )}
    </section>
  );
}

export default UserManagement;
