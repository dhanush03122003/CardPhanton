import { useEffect, useState } from "react";
import { apiRequest } from "../api";
import { useToast } from "../components/Toast";

function maskCardNumber(pan) {
  const value = String(pan ?? "");
  if (value.length <= 4) return value || "-";
  return `•••• •••• •••• ${value.slice(-4)}`;
}

function formatTimestamp(value) {
  if (!value) return "-";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "-" : date.toLocaleString();
}

function Consents() {
  const [cards, setCards] = useState([]);
  const [loading, setLoading] = useState(true);
  const [consentingCardId, setConsentingCardId] = useState(null);
  const showToast = useToast();

  const loadConsents = async () => {
    setLoading(true);
    try {
      const data = await apiRequest("/api/auth/google/consents");
      setCards(Array.isArray(data?.cards) ? data.cards : []);
    } catch (error) {
      showToast(error.message, "error");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadConsents();
  }, []);

  const requestConsent = async (cardId) => {
    setConsentingCardId(cardId);
    try {
      const data = await apiRequest(`/api/auth/google/consent/${cardId}`);
      window.location.assign(data.auth_url);
    } catch (error) {
      setConsentingCardId(null);
      showToast(
        error.code === "CARD_EMAIL_REQUIRED"
          ? "Add an email address to this card before requesting consent."
          : error.message,
        "error",
      );
    }
  };

  const revokeConsent = async (cardId) => {
    setConsentingCardId(cardId);
    try {
      await apiRequest(`/api/auth/google/consent/${cardId}`, {
        method: "DELETE",
      });
      setCards((current) =>
        current.map((card) =>
          card.id === cardId
            ? { ...card, gmail: null, last_updated: null }
            : card,
        ),
      );
      showToast("Google consent revoked.", "success");
    } catch (error) {
      showToast(error.message, "error");
    } finally {
      setConsentingCardId(null);
    }
  };

  return (
    <section className='card'>
      <div className='mb-6 border-b border-gray-100 pb-4'>
        <h2 className='font-semibold text-gray-900'>Google consents</h2>
        <p className='mt-1 text-sm text-gray-500'>
          Connect a Gmail account to each card used for email-based refreshes.
        </p>
      </div>

      {loading ? (
        <div className='h-24 animate-pulse rounded-lg bg-gray-50' />
      ) : cards.length === 0 ? (
        <p className='py-12 text-center text-sm text-gray-500'>
          No cards found.
        </p>
      ) : (
        <div className='overflow-x-auto'>
          <table className='w-full min-w-[680px] text-left text-sm'>
            <thead className='border-b border-gray-200 text-xs uppercase tracking-wide text-gray-500'>
              <tr>
                <th className='px-3 py-3 font-medium'>Card number</th>
                <th className='px-3 py-3 font-medium'>Gmail</th>
                <th className='px-3 py-3 font-medium'>Consent given</th>
                <th className='px-3 py-3 text-right font-medium'>Action</th>
              </tr>
            </thead>
            <tbody className='divide-y divide-gray-100'>
              {cards.map((card) => {
                const hasConsent = Boolean(card.gmail && card.last_updated);
                return (
                  <tr key={card.id}>
                    <td className='whitespace-nowrap px-3 py-4 font-medium text-gray-900'>
                      {maskCardNumber(card.pan)}
                    </td>
                    <td className='px-3 py-4 text-gray-600'>
                      {card.gmail || "-"}
                    </td>
                    <td className='whitespace-nowrap px-3 py-4 text-gray-600'>
                      {formatTimestamp(card.last_updated)}
                    </td>
                    <td className='px-3 py-4 text-right'>
                      {hasConsent ? (
                        <button
                          type='button'
                          className='btn-secondary text-red-600'
                          disabled={consentingCardId === card.id}
                          onClick={() => revokeConsent(card.id)}
                        >
                          {consentingCardId === card.id
                            ? "Revoking..."
                            : "Revoke"}
                        </button>
                      ) : (
                        <button
                          type='button'
                          className='btn-primary'
                          disabled={consentingCardId === card.id}
                          onClick={() => requestConsent(card.id)}
                        >
                          {consentingCardId === card.id
                            ? "Opening..."
                            : "Get consent"}
                        </button>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}

export default Consents;
