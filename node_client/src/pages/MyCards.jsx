import { useEffect, useState } from "react";
import { apiRequest } from "../api";
import { useToast } from "../components/Toast";

const emptyCard = {
  pan: "",
  cardholder_name: "",
  bank_name: "",
  product_name: "",
  payment_method_type: "Credit",
  card_brand: "Visa",
  linked_phone_number: "",
  exp_month: "",
  exp_year: "",
  cvv: "",
};

function MyCards() {
  const [cards, setCards] = useState([]);
  const [form, setForm] = useState(emptyCard);
  const [editingId, setEditingId] = useState(null);
  const [formOpen, setFormOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const showToast = useToast();

  const loadCards = async () => {
    setLoading(true);
    try {
      const data = await apiRequest("/api/auth/cards");
      setCards(Array.isArray(data?.cards) ? data.cards : []);
    } catch (requestError) {
      setError(requestError.message);
      showToast(requestError.message, "error");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadCards();
  }, []);
  const updateField = (event) =>
    setForm((current) => ({
      ...current,
      [event.target.name]: event.target.value,
    }));
  const closeForm = () => {
    setFormOpen(false);
    setEditingId(null);
    setForm(emptyCard);
  };
  const openCreate = () => {
    setForm(emptyCard);
    setEditingId(null);
    setError("");
    setFormOpen(true);
  };
  const openEdit = (card) => {
    setEditingId(card.id);
    setForm({ ...emptyCard, ...card });
    setError("");
    setFormOpen(true);
  };

  const submit = async (event) => {
    event.preventDefault();
    setSaving(true);
    setError("");
    try {
      await apiRequest(
        editingId ? `/api/auth/cards/${editingId}` : "/api/auth/cards",
        {
          method: editingId ? "PUT" : "POST",
          body: JSON.stringify({
            ...form,
            exp_month: Number(form.exp_month),
            exp_year: Number(form.exp_year),
            cvv: Number(form.cvv),
          }),
        },
      );
      await loadCards();
      showToast(
        editingId ? "Card updated successfully." : "Card created successfully.",
        "success",
      );
      closeForm();
    } catch (requestError) {
      setError(requestError.message);
      showToast(requestError.message, "error");
    } finally {
      setSaving(false);
    }
  };

  const deleteCard = async (card) => {
    if (!window.confirm(`Delete ${card.product_name || "this card"}?`)) return;
    try {
      await apiRequest(`/api/auth/cards/${card.id}`, { method: "DELETE" });
      setCards((current) => current.filter((item) => item.id !== card.id));
      showToast("Card deleted successfully.", "success");
    } catch (requestError) {
      setError(requestError.message);
      showToast(requestError.message, "error");
    }
  };

  const fields = [
    ["pan", "PAN", "text", "16 digits"],
    ["cardholder_name", "Cardholder Name", "text", "Name on card"],
    ["bank_name", "Bank", "text", "Axis Bank"],
    ["product_name", "Product Name", "text", "Flipkart Axis"],
    ["linked_phone_number", "Linked Phone", "tel", "10-digit mobile number"],
    ["exp_month", "Expiry Month", "number", "MM"],
    ["exp_year", "Expiry Year", "number", "YYYY"],
    ["cvv", "CVV", "password", "3 digits"],
  ];

  return (
    <section className='card'>
      <div className='mb-6 flex items-center justify-between border-b border-gray-100 pb-4 dark:border-slate-700'>
        <div>
          <h2 className='font-semibold text-gray-900 dark:text-white'>
            My Cards
          </h2>
          <p className='mt-1 text-sm text-gray-500'>
            Manage cards linked to your account.
          </p>
        </div>
        <button type='button' onClick={openCreate} className='btn-primary'>
          Create New Card
        </button>
      </div>
      {error && (
        <p className='mb-4 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700 dark:bg-red-950 dark:text-red-300'>
          {error}
        </p>
      )}
      {loading ? (
        <div className='h-24 animate-pulse rounded-lg bg-gray-50 dark:bg-slate-800' />
      ) : cards.length === 0 ? (
        <div className='py-12 text-center text-sm text-gray-500'>
          <p className='font-medium'>No cards added yet.</p>
          <button
            type='button'
            onClick={openCreate}
            className='mt-3 text-sm font-semibold text-blue-600 hover:text-blue-700'
          >
            Add your first card
          </button>
        </div>
      ) : (
        <div className='overflow-x-auto rounded-xl border border-gray-200 dark:border-slate-700'>
          <table className='w-full text-left text-sm'>
            <thead>
              <tr className='sticky top-0 z-10 border-b border-gray-200 bg-white text-xs uppercase tracking-wide text-gray-500 dark:border-slate-700 dark:bg-slate-900'>
                <th className='px-3 py-3'>Bank / Product</th>
                <th className='px-3 py-3'>Cardholder</th>
                <th className='px-3 py-3'>Last 4</th>
                <th className='px-3 py-3'>Expiry</th>
                <th className='px-3 py-3'>Brand</th>
                <th className='px-3 py-3 text-right'>Actions</th>
              </tr>
            </thead>
            <tbody>
              {cards.map((card) => (
                <tr
                  key={card.id}
                  className='border-b border-gray-100 transition-colors hover:bg-slate-50 dark:border-slate-800 dark:hover:bg-slate-800/70'
                >
                  <td className='px-3 py-4 font-medium text-gray-900 dark:text-white'>
                    {card.bank_name}
                    <div className='text-xs text-gray-500'>
                      {card.product_name || "Standard card"}
                    </div>
                  </td>
                  <td className='px-3 py-4 text-gray-700 dark:text-slate-300'>
                    {card.cardholder_name}
                  </td>
                  <td className='px-3 py-4 font-mono text-gray-700 dark:text-slate-300'>
                    •••• {String(card.pan).slice(-4)}
                  </td>
                  <td className='px-3 py-4 text-gray-700 dark:text-slate-300'>
                    {card.exp_month}/{card.exp_year}
                  </td>
                  <td className='px-3 py-4 text-gray-700 dark:text-slate-300'>
                    {card.card_brand}
                  </td>
                  <td className='px-3 py-4 text-right'>
                    <button
                      type='button'
                      onClick={() => openEdit(card)}
                      className='mr-2 min-h-11 px-2 text-xs font-medium text-slate-600 hover:text-slate-900 dark:text-slate-300'
                    >
                      Update
                    </button>
                    <button
                      type='button'
                      onClick={() => deleteCard(card)}
                      className='min-h-11 px-2 text-xs font-medium text-red-600'
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      {formOpen && (
        <div
          className='fixed inset-0 z-50 flex items-end justify-center bg-black/40 p-0 backdrop-blur-sm md:items-center md:p-4'
          onClick={closeForm}
        >
          <form
            onSubmit={submit}
            onClick={(event) => event.stopPropagation()}
            className='max-h-[95vh] w-full max-w-2xl overflow-y-auto rounded-t-2xl bg-white p-4 shadow-xl dark:bg-slate-900 sm:p-6 md:max-h-[90vh] md:rounded-xl'
          >
            <div className='mb-5 flex items-center justify-between'>
              <h3 className='text-lg font-semibold text-gray-900 dark:text-white'>
                {editingId ? "Update Card" : "Create New Card"}
              </h3>
              <button
                type='button'
                onClick={closeForm}
                aria-label='Close card form'
                className='flex h-11 w-11 items-center justify-center rounded-full text-2xl text-gray-500 hover:bg-gray-100 hover:text-gray-900 dark:hover:bg-slate-800 dark:hover:text-white'
              >
                ×
              </button>
            </div>
            <div className='grid grid-cols-1 gap-4 sm:grid-cols-2'>
              {fields.map(([name, label, type, placeholder]) => (
                <label
                  key={name}
                  className='text-sm font-medium text-gray-700 dark:text-slate-300'
                >
                  {label}
                  <input
                    name={name}
                    type={type}
                    value={form[name] || ""}
                    onChange={updateField}
                    placeholder={placeholder}
                    required
                    className='input-field mt-1'
                  />
                </label>
              ))}
              <label className='text-sm font-medium text-gray-700 dark:text-slate-300'>
                Payment Type
                <select
                  name='payment_method_type'
                  value={form.payment_method_type}
                  onChange={updateField}
                  className='input-field mt-1'
                >
                  <option>Credit</option>
                  <option>Debit</option>
                </select>
              </label>
              <label className='text-sm font-medium text-gray-700 dark:text-slate-300'>
                Brand
                <select
                  name='card_brand'
                  value={form.card_brand}
                  onChange={updateField}
                  className='input-field mt-1'
                >
                  <option>Visa</option>
                  <option>Mastercard</option>
                  <option>American Express</option>
                  <option>Discover</option>
                  <option>RuPay</option>
                  <option>Other</option>
                </select>
              </label>
            </div>
            <button disabled={saving} className='btn-primary mt-6 w-full'>
              {saving ? "Saving..." : "Save Card"}
            </button>
          </form>
        </div>
      )}
    </section>
  );
}

export default MyCards;
