import { useEffect, useState } from "react";
import { startRegistration } from "@simplewebauthn/browser";
import GlobalCards from "./GlobalCards";
import MyCards from "./MyCards";
import UserManagement from "./UserManagement";

function Dashboard({ user }) {
  const [authenticators, setAuthenticators] = useState([]);
  const [aaguidMap, setAaguidMap] = useState({});
  const [loading, setLoading] = useState(true);
  const [registering, setRegistering] = useState(false);
  const [isAdmin, setIsAdmin] = useState(false);
  const [activeTab, setActiveTab] = useState("security");

  const fetchAuthenticators = async () => {
    try {
      const response = await fetch("/api/auth/me", { credentials: "include" });
      if (response.ok) {
        const data = await response.json();
        setAuthenticators(data.authenticators || []);
      } else if (response.status === 401 || response.status === 404) {
        window.location.reload();
      }
    } catch (error) {
      console.error("Error fetching authenticators:", error);
    }
  };

  useEffect(() => {
    fetchAuthenticators().finally(() => setLoading(false));
    fetch(
      "https://raw.githubusercontent.com/passkeydeveloper/passkey-authenticator-aaguids/main/aaguid.json",
    )
      .then((response) => (response.ok ? response.json() : {}))
      .then(setAaguidMap)
      .catch((error) => console.error("Error fetching AAGUID map:", error));
    fetch("/api/admin/status", { credentials: "include" })
      .then((response) => (response.ok ? response.json() : { isAdmin: false }))
      .then((data) => setIsAdmin(data.isAdmin === true))
      .catch((error) => console.error("Error fetching admin status:", error));
  }, []);

  const getDeviceInfo = (auth) => {
    const attachmentType = auth.attachmentType || auth.attachment_type;
    const transports = auth.transports || [];
    const aaguid = (auth.aaguid || "").toLowerCase();
    const brand = aaguidMap[aaguid];
    if (brand) {
      return {
        name: brand.name,
        logo: brand.icon_light || brand.icon_dark,
        description:
          attachmentType === "platform" ? "Platform passkey" : "Security key",
      };
    }
    if (
      transports.includes("internal") &&
      /Mac|iPhone|iPad/i.test(navigator.userAgent)
    ) {
      return {
        name: "Apple iCloud Keychain",
        logo: null,
        description: "Face ID / Touch ID",
      };
    }
    return {
      name:
        attachmentType === "platform" ? "Built-in Biometrics" : "Security Key",
      logo: null,
      description: "Standard WebAuthn device",
    };
  };

  const handleRegisterNewDevice = async () => {
    const nickname = window.prompt(
      "Enter a label for this passkey:",
      "My Personal Device",
    );
    if (nickname === null) return;
    setRegistering(true);
    try {
      const optionsResponse = await fetch(
        "/api/auth/generate-additional-device-options",
        { credentials: "include" },
      );
      const optionsData = await optionsResponse.json();
      if (!optionsResponse.ok)
        throw new Error(optionsData.error || "Failed to generate options.");
      const registrationResult = await startRegistration({
        optionsJSON: optionsData.publicKey,
      });
      const verifyResponse = await fetch("/api/auth/verify-registration", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          username: user.username,
          verification: registrationResult,
          nickname: nickname.trim() || "Unnamed Passkey",
        }),
      });
      const verifyData = await verifyResponse.json();
      if (!verifyResponse.ok)
        throw new Error(verifyData.error || "Passkey registration failed.");
      await fetchAuthenticators();
    } catch (error) {
      window.alert(
        error.name === "NotAllowedError"
          ? "Registration timed out or cancelled."
          : error.message,
      );
    } finally {
      setRegistering(false);
    }
  };

  const editNickname = async (auth) => {
    const nickname = window.prompt(
      "Edit the nickname for this device:",
      auth.nickname || "",
    );
    if (nickname === null || nickname.trim() === (auth.nickname || "")) return;
    const response = await fetch(
      `/api/auth/authenticator/${auth.id}/nickname`,
      {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ nickname: nickname.trim() }),
      },
    );
    const data = await response.json();
    if (!response.ok) window.alert(data.error || "Failed to update nickname.");
    else await fetchAuthenticators();
  };

  const deleteAuthenticator = async (auth) => {
    if (authenticators.length <= 1)
      return window.alert("You must keep at least one passkey.");
    if (!window.confirm(`Delete ${auth.nickname || "this passkey"}?`)) return;
    const response = await fetch(`/api/auth/authenticator/${auth.id}`, {
      method: "DELETE",
      credentials: "include",
    });
    const data = await response.json();
    if (!response.ok) window.alert(data.error || "Failed to delete passkey.");
    else await fetchAuthenticators();
  };

  return (
    <div className='min-h-[calc(100vh-4rem)] p-4 sm:p-8'>
      <div className='mx-auto max-w-5xl space-y-8'>
        <header className='flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between'>
          <div>
            <h1 className='text-2xl font-bold text-gray-900'>
              Security Details
            </h1>
            <p className='mt-1 text-sm text-gray-500'>
              Manage your passkeys and active authentication methods.
            </p>
          </div>
          <button
            onClick={handleRegisterNewDevice}
            disabled={registering || loading}
            className='btn-primary'
          >
            {registering ? "Adding..." : "+ Add Passkey"}
          </button>
        </header>
        <nav
          className='flex gap-1 border-b border-gray-200'
          aria-label='Dashboard sections'
        >
          <button
            onClick={() => setActiveTab("security")}
            className={`border-b-2 px-4 py-2.5 text-sm font-medium ${activeTab === "security" ? "border-slate-900 text-slate-900" : "border-transparent text-gray-500"}`}
          >
            Security
          </button>
          <button
            onClick={() => setActiveTab("my-cards")}
            className={`border-b-2 px-4 py-2.5 text-sm font-medium ${activeTab === "my-cards" ? "border-slate-900 text-slate-900" : "border-transparent text-gray-500"}`}
          >
            My Cards
          </button>
          {isAdmin && (
            <button
              onClick={() => setActiveTab("management")}
              className={`border-b-2 px-4 py-2.5 text-sm font-medium ${activeTab === "management" ? "border-slate-900 text-slate-900" : "border-transparent text-gray-500"}`}
            >
              User Management
            </button>
          )}
          {isAdmin && (
            <button
              onClick={() => setActiveTab("global-cards")}
              className={`border-b-2 px-4 py-2.5 text-sm font-medium ${activeTab === "global-cards" ? "border-slate-900 text-slate-900" : "border-transparent text-gray-500"}`}
            >
              Global Cards
            </button>
          )}
        </nav>
        {activeTab === "my-cards" ? (
          <MyCards />
        ) : activeTab === "global-cards" && isAdmin ? (
          <GlobalCards />
        ) : activeTab === "management" && isAdmin ? (
          <UserManagement />
        ) : (
          <section className='card'>
            <div className='mb-6 flex items-center justify-between border-b border-gray-100 pb-4'>
              <div>
                <h2 className='font-semibold text-gray-900'>Your passkeys</h2>
                <p className='mt-1 text-sm text-gray-500'>
                  Passkeys allow you to securely log in without a password.
                </p>
              </div>
              <span className='rounded bg-gray-100 px-2 py-1 text-xs font-semibold text-gray-700'>
                {authenticators.length}
              </span>
            </div>
            {loading ? (
              <div className='h-20 animate-pulse rounded-lg bg-gray-50' />
            ) : authenticators.length === 0 ? (
              <p className='py-12 text-center text-sm text-gray-500'>
                No passkeys configured.
              </p>
            ) : (
              <div className='space-y-3'>
                {authenticators.map((auth) => {
                  const info = getDeviceInfo(auth);
                  return (
                    <div
                      key={auth.id}
                      className='flex items-center justify-between rounded-lg border border-gray-200 p-4'
                    >
                      <div className='flex items-center gap-3'>
                        <div className='flex h-10 w-10 items-center justify-center rounded-lg border border-gray-100 bg-gray-50'>
                          {info.logo ? (
                            <img
                              src={info.logo}
                              alt=''
                              className='h-6 w-6 object-contain'
                            />
                          ) : (
                            <span className='text-lg'>
                              {info.name.includes("Apple")
                                ? "🍏"
                                : info.name.includes("Biometric")
                                  ? "👤"
                                  : "🔑"}
                            </span>
                          )}
                        </div>
                        <div>
                          <h3 className='text-sm font-semibold text-gray-900'>
                            {auth.nickname || "Unnamed Passkey"}
                          </h3>
                          <p className='mt-1 text-xs text-gray-500'>
                            {info.name} · {info.description}
                          </p>
                        </div>
                      </div>
                      <div className='flex gap-2'>
                        <button
                          onClick={() => editNickname(auth)}
                          className='btn-secondary'
                        >
                          Rename
                        </button>
                        <button
                          onClick={() => deleteAuthenticator(auth)}
                          className='btn-secondary text-red-600'
                        >
                          Delete
                        </button>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </section>
        )}
      </div>
    </div>
  );
}

export default Dashboard;
