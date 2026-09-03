import { useEffect, useState } from "react";
import { apiRequest } from "../api";

const formatPan = (pan = "") => pan.replace(/(.{4})/g, "$1 ").trim();

function GlobalCards() {
  const [cards, setCards] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    apiRequest("/api/cards/global")
      .then((data) => setCards(Array.isArray(data?.cards) ? data.cards : []))
      .catch((requestError) => setError(requestError.message))
      .finally(() => setLoading(false));
  }, []);

  const cardTheme = (bank = "") => {
    if (/axis/i.test(bank)) return "from-red-900 via-red-800 to-slate-950";
    if (/sbi/i.test(bank)) return "from-blue-900 via-blue-800 to-slate-950";
    if (/hdfc/i.test(bank)) return "from-indigo-900 via-blue-900 to-slate-950";
    return "from-slate-700 via-slate-800 to-slate-950";
  };

  return (
    <section className='card'>
      <div className='mb-6 border-b border-gray-100 pb-4 dark:border-slate-700'>
        <h2 className='font-semibold text-gray-900 dark:text-white'>
          Global Cards
        </h2>
        <p className='mt-1 text-sm text-gray-500'>
          Admin view of all cards in the system.
        </p>
      </div>
      {loading && (
        <div className='h-56 animate-pulse rounded-2xl bg-gray-100 dark:bg-slate-800' />
      )}
      {error && (
        <p className='rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700 dark:bg-red-950 dark:text-red-300'>
          {error}
        </p>
      )}
      {!loading && !error && cards.length === 0 && (
        <p className='py-12 text-center text-sm text-gray-500'>
          No cards found.
        </p>
      )}
      {!loading && !error && cards.length > 0 && (
        <div className='grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3'>
          {cards.map((card) => (
            <div
              key={card.id}
              className={`aspect-[1.7/1] w-full max-w-sm rounded-2xl bg-gradient-to-br ${cardTheme(card.bank_name)} p-6 text-white shadow-xl`}
            >
              <div className='flex items-start justify-between'>
                <div className='h-10 w-14 rounded-md border border-amber-200/50 bg-gradient-to-br from-amber-200 to-amber-500 shadow-inner'>
                  <div className='mt-2 h-1 border-y border-amber-700/40' />
                  <div className='mt-1 h-1 border-y border-amber-700/40' />
                </div>
                <div className='text-right'>
                  <p className='text-sm font-semibold'>{card.bank_name}</p>
                  <p className='max-w-32 text-xs text-white/70'>
                    {card.product_name || "Card"}
                  </p>
                </div>
              </div>
              <p className='mt-7 font-mono text-lg tracking-[0.18em] sm:text-xl'>
                {formatPan(card.pan)}
              </p>
              <div className='mt-5 flex items-end justify-between'>
                <div>
                  <p className='text-[9px] uppercase text-white/60'>
                    Cardholder
                  </p>
                  <p className='text-sm font-medium uppercase'>
                    {card.cardholder_name}
                  </p>
                </div>
                <div>
                  <p className='text-[9px] uppercase text-white/60'>Expires</p>
                  <p className='text-sm'>
                    {String(card.exp_month).padStart(2, "0")}/{card.exp_year}
                  </p>
                </div>
                <p className='text-lg font-bold italic'>{card.card_brand}</p>
              </div>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}

export default GlobalCards;
