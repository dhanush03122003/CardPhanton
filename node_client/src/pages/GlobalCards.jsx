import { useEffect, useState } from "react";
import { apiRequest } from "../api";
import { useToast } from "../components/Toast";

const formatPan = (pan = "") => pan.replace(/(.{4})/g, "$1 ").trim();

function PhysicalCard({
  card,
  theme,
  copied,
  onCopy,
  onOpen,
  expanded = false,
}) {
  return (
    <div
      role={onOpen ? "button" : undefined}
      tabIndex={onOpen ? 0 : undefined}
      onClick={onOpen}
      onKeyDown={(event) => {
        if (onOpen && (event.key === "Enter" || event.key === " ")) {
          event.preventDefault();
          onOpen();
        }
      }}
      className={`relative w-full ${expanded ? "aspect-[1.58/1] max-w-[620px]" : "min-h-[200px] max-w-[340px] cursor-pointer transition-transform duration-200 hover:-translate-y-1 hover:shadow-[0_24px_55px_rgba(15,23,42,0.35)] sm:min-h-[235px] sm:max-w-[380px]"} overflow-hidden rounded-2xl border border-white/10 bg-gradient-to-br ${theme} p-4 text-white shadow-2xl sm:p-6`}
    >
      <div className='flex items-start justify-between'>
        <div>
          <div className='h-10 w-14 rounded-md border border-amber-200/50 bg-gradient-to-br from-amber-200 to-amber-500 shadow-inner'>
            <div className='mt-2 h-1 border-y border-amber-700/40' />
            <div className='mt-1 h-1 border-y border-amber-700/40' />
          </div>
          <span className='mt-2 inline-flex rounded-full border border-white/20 px-2 py-0.5 text-[10px] uppercase tracking-widest text-white/70 backdrop-blur-sm'>
            {card.payment_method_type}
          </span>
        </div>
        <div className='text-right'>
          <p className='max-w-40 truncate text-sm font-semibold'>
            {card.bank_name}
          </p>
          <p className='max-w-40 truncate text-xs text-white/70'>
            {card.product_name || "Card"}
          </p>
        </div>
      </div>
      <div className='mt-5 flex items-center gap-1 sm:mt-6 sm:gap-2'>
        <p className='whitespace-nowrap font-mono text-sm tracking-[0.12em] text-white/90 sm:text-xl sm:tracking-widest'>
          {formatPan(card.pan)}
        </p>
        <button
          type='button'
          onClick={(event) => {
            event.stopPropagation();
            onCopy(card);
          }}
          className='shrink-0 text-white/50 transition-colors hover:text-white'
          aria-label={copied ? "Copied" : "Copy card number"}
          title={copied ? "Copied" : "Copy card number"}
        >
          {copied ? (
            <svg
              className='h-5 w-5 text-emerald-300'
              viewBox='0 0 24 24'
              fill='none'
              stroke='currentColor'
              strokeWidth='2.5'
            >
              <path d='m5 12 4 4L19 6' />
            </svg>
          ) : (
            <svg
              className='h-5 w-5'
              viewBox='0 0 24 24'
              fill='none'
              stroke='currentColor'
              strokeWidth='2'
            >
              <rect x='9' y='9' width='11' height='11' rx='2' />
              <path d='M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1' />
            </svg>
          )}
        </button>
      </div>
      <div className='absolute inset-x-4 bottom-4 flex items-end justify-between gap-2 sm:inset-x-6 sm:bottom-6 sm:gap-3'>
        <div className='min-w-0'>
          <p className='text-[8px] uppercase text-white/50'>Cardholder</p>
          <p className='truncate text-sm font-medium uppercase text-white'>
            {card.cardholder_name}
          </p>
        </div>
        <div className='flex shrink-0 items-end gap-3'>
          <div>
            <p className='text-[8px] uppercase text-white/50'>Valid Thru</p>
            <p className='text-sm text-white'>
              {String(card.exp_month).padStart(2, "0")}/
              {String(card.exp_year).slice(-2)}
            </p>
          </div>
          <div>
            <p className='text-[8px] uppercase text-white/50'>CVV</p>
            <p className='text-sm text-white'>{card.cvv}</p>
          </div>
          <p className='text-lg font-bold italic'>{card.card_brand}</p>
        </div>
      </div>
    </div>
  );
}

function GlobalCards() {
  const [cards, setCards] = useState([]);
  const [copiedCardId, setCopiedCardId] = useState(null);
  const [selectedCard, setSelectedCard] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [filters, setFilters] = useState({
    search: "",
    type: "ALL",
    brand: "ALL",
    bank: "ALL",
  });
  const showToast = useToast();

  const filteredCards = cards.filter((card) => {
    const search = filters.search.trim().toLowerCase();
    const searchable = [
      card.bank_name,
      card.product_name,
      card.cardholder_name,
      card.pan,
      card.card_brand,
    ]
      .join(" ")
      .toLowerCase();
    return (
      (!search || searchable.includes(search)) &&
      (filters.type === "ALL" || card.payment_method_type === filters.type) &&
      (filters.brand === "ALL" || card.card_brand === filters.brand) &&
      (filters.bank === "ALL" || card.bank_name === filters.bank)
    );
  });

  const banks = [
    ...new Set(cards.map((card) => card.bank_name).filter(Boolean)),
  ].sort();
  const brands = [
    ...new Set(cards.map((card) => card.card_brand).filter(Boolean)),
  ].sort();

  const updateFilter = (event) => {
    const { name, value } = event.target;
    setFilters((current) => ({ ...current, [name]: value }));
  };

  useEffect(() => {
    apiRequest("/api/cards")
      .then((data) => setCards(Array.isArray(data?.cards) ? data.cards : []))
      .catch((requestError) => {
        setError(requestError.message);
        showToast(requestError.message, "error");
      })
      .finally(() => setLoading(false));
  }, []);

  const cardTheme = (bank = "") => {
    if (/axis/i.test(bank)) return "from-red-900 via-red-800 to-slate-950";
    if (/sbi/i.test(bank)) return "from-blue-900 via-blue-800 to-slate-950";
    if (/hdfc/i.test(bank)) return "from-indigo-900 via-blue-900 to-slate-950";
    return "from-slate-700 via-slate-800 to-slate-950";
  };

  const copyPan = async (card) => {
    try {
      await navigator.clipboard.writeText(card.pan);
      setCopiedCardId(card.id);
      window.setTimeout(() => setCopiedCardId(null), 2000);
    } catch (copyError) {
      console.error("Unable to copy card number:", copyError);
    }
  };

  return (
    <section className='card'>
      <div className='mb-6 border-b border-gray-100 pb-4 dark:border-slate-700'>
        <h2 className='font-semibold text-gray-900 dark:text-white'>
          All Cards
        </h2>
        <p className='mt-1 text-sm text-gray-500'>
          View all cards available in the system.
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
      {!loading && !error && cards.length > 0 && (
        <div className='mb-6 rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-slate-700 dark:bg-slate-900'>
          <div className='mb-3 flex items-center justify-between'>
            <h3 className='text-xs font-semibold uppercase tracking-wider text-slate-500'>
              Filter cards
            </h3>
            {(filters.search ||
              filters.type !== "ALL" ||
              filters.brand !== "ALL" ||
              filters.bank !== "ALL") && (
              <button
                type='button'
                onClick={() =>
                  setFilters({
                    search: "",
                    type: "ALL",
                    brand: "ALL",
                    bank: "ALL",
                  })
                }
                className='text-xs font-semibold text-blue-600 hover:text-blue-700'
              >
                Clear filters
              </button>
            )}
          </div>
          <div className='grid grid-cols-1 gap-3 md:grid-cols-4'>
            <input
              name='search'
              value={filters.search}
              onChange={updateFilter}
              placeholder='Search bank, product, holder, PAN'
              className='input-field md:col-span-1'
            />
            <select
              name='type'
              value={filters.type}
              onChange={updateFilter}
              className='input-field'
            >
              <option value='ALL'>All types</option>
              <option value='Credit'>Credit</option>
              <option value='Debit'>Debit</option>
            </select>
            <select
              name='brand'
              value={filters.brand}
              onChange={updateFilter}
              className='input-field'
            >
              <option value='ALL'>All brands</option>
              {brands.map((brand) => (
                <option key={brand} value={brand}>
                  {brand}
                </option>
              ))}
            </select>
            <select
              name='bank'
              value={filters.bank}
              onChange={updateFilter}
              className='input-field'
            >
              <option value='ALL'>All banks</option>
              {banks.map((bank) => (
                <option key={bank} value={bank}>
                  {bank}
                </option>
              ))}
            </select>
          </div>
          <p className='mt-3 text-xs text-slate-500'>
            {filteredCards.length} result{filteredCards.length === 1 ? "" : "s"}{" "}
            shown
          </p>
        </div>
      )}
      {!loading && !error && cards.length === 0 && (
        <p className='py-12 text-center text-sm text-gray-500'>
          No cards found.
        </p>
      )}
      {!loading && !error && cards.length > 0 && filteredCards.length === 0 && (
        <div className='rounded-xl border border-dashed border-gray-200 py-12 text-center dark:border-slate-700'>
          <p className='text-sm font-medium text-gray-700 dark:text-slate-300'>
            No cards match these filters.
          </p>
          <button
            type='button'
            onClick={() =>
              setFilters({ search: "", type: "ALL", brand: "ALL", bank: "ALL" })
            }
            className='mt-3 text-sm font-semibold text-blue-600 hover:text-blue-700'
          >
            Clear filters
          </button>
        </div>
      )}
      {!loading && !error && filteredCards.length > 0 && (
        <div className='grid grid-cols-1 justify-items-center gap-8 md:grid-cols-2 2xl:grid-cols-3'>
          {filteredCards.map((card) => (
            <PhysicalCard
              key={card.id}
              card={card}
              theme={cardTheme(card.bank_name)}
              copied={copiedCardId === card.id}
              onCopy={copyPan}
              onOpen={() => setSelectedCard(card)}
            />
          ))}
        </div>
      )}
      {selectedCard && (
        <div
          className='fixed inset-0 z-50 flex items-center justify-center bg-slate-950/70 p-4 backdrop-blur-sm'
          onClick={() => setSelectedCard(null)}
        >
          <div
            className='w-full max-w-[680px]'
            onClick={(event) => event.stopPropagation()}
          >
            <div className='mb-3 flex justify-end'>
              <button
                type='button'
                onClick={() => setSelectedCard(null)}
                className='rounded-full border border-white/20 px-3 py-1 text-sm text-white/70 hover:text-white'
              >
                Close
              </button>
            </div>
            <PhysicalCard
              card={selectedCard}
              theme={cardTheme(selectedCard.bank_name)}
              copied={copiedCardId === selectedCard.id}
              onCopy={copyPan}
              expanded
            />
          </div>
        </div>
      )}
    </section>
  );
}

export default GlobalCards;
