import { useEffect, useState } from "react";
import { apiRequest } from "../api";

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

  const loadCards = async () => {
    setLoading(true);
    try {
      const data = await apiRequest("/api/auth/cards");
      setCards(Array.isArray(data?.cards) ? data.cards : []);
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadCards();
  }, []);

  const updateField = (event) => {
    const { name, value } = event.target;
    setForm((current) => ({ ...current, [name]: value }));
  };

  const openCreate = () => {
    setFormOpen(true);
    setEditingId(null);
    setForm(emptyCard);
    setError("");
  };

  const openEdit = (card) => {
    setFormOpen(true);
    setEditingId(card.id);
    setForm({
      pan: card.pan || "",
      cardholder_name: card.cardholder_name || "",
      bank_name: card.bank_name || "",
      product_name: card.product_name || "",
      payment_method_type: card.payment_method_type || "Credit",
      card_brand: card.card_brand || "Visa",
      linked_phone_number: card.linked_phone_number || "",
      exp_month: card.exp_month || "",
      exp_year: card.exp_year || "",
      cvv: card.cvv || "",
    });
    setError("");
  };

  const submit = async (event) => {
    event.preventDefault();
    setSaving(true);
    setError("");
    try {
      const path = editingId
        ? `/api/auth/cards/${editingId}`
        : "/api/auth/cards";
      await apiRequest(path, {
        method: editingId ? "PUT" : "POST",
        body: JSON.stringify({
          ...form,
          exp_month: Number(form.exp_month),
          exp_year: Number(form.exp_year),
          cvv: Number(form.cvv),
        }),
      });
      await loadCards();
      setFormOpen(false);
      setEditingId(null);
      setForm(emptyCard);
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setSaving(false);
    }
  };

  const deleteCard = async (card) => {
    if (!window.confirm(`Delete ${card.product_name || "this card"}?`)) return;
    try {
      await apiRequest(`/api/auth/cards/${card.id}`, { method: "DELETE" });
      setCards((current) => current.filter((item) => item.id !== card.id));
    } catch (requestError) {
      setError(requestError.message);
    }
  };

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
        <p className='py-12 text-center text-sm text-gray-500'>
          No cards added yet.
        </p>
      ) : (
        <div className='overflow-x-auto'>
          <table className='w-full text-left text-sm'>
            <thead>
              <tr className='border-b border-gray-200 text-xs uppercase tracking-wide text-gray-500 dark:border-slate-700'>
                <th className='px-3 py-3'>Bank / Product</th>
                <th className='px-3 py-3'>Cardholder</th>
                <th className='px-3 py-3'>Last 4</th>
                <th className='px-3 py-3'>Expiry</th>
                <th className='px-3 py-3'>Brand</th>
                <th className='px-3 py-3 text-right'>Actions</th>
              </tr>
            </thead>
            <tbody className='divide-y divide-gray-100'>
              {cards.map((card) => (
                <tr key={card.id}>
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
                      className='mr-3 text-xs font-medium text-slate-600 hover:text-slate-900 dark:text-slate-300'
                    >
                      Update
                    </button>
                    <button
                      type='button'
                      onClick={() => deleteCard(card)}
                      className='text-xs font-medium text-red-600'
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
          className='fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4'
          onClick={() => {
            setFormOpen(false);
            setEditingId(null);
            setForm(emptyCard);
          }}
        >
          <form
            onSubmit={submit}
            onClick={(event) => event.stopPropagation()}
            className='max-h-[90vh] w-full max-w-2xl overflow-y-auto rounded-xl bg-white p-6 shadow-xl dark:bg-slate-900'
          >
            <div className='mb-5 flex items-center justify-between'>
              <h3 className='text-lg font-semibold text-gray-900 dark:text-white'>
                {editingId ? "Update Card" : "Create New Card"}
              </h3>
              <button
                type='button'
                onClick={() => {
                  setFormOpen(false);
                  setEditingId(null);
                  setForm(emptyCard);
                }}
                className='text-sm text-gray-500'
              >
                Close
              </button>
            </div>
            <div className='grid grid-cols-1 gap-4 sm:grid-cols-2'>
              {[
                ["pan", "PAN", "text", "16 digits"],
                ["cardholder_name", "Cardholder Name", "text", "Name on card"],
                ["bank_name", "Bank", "text", "Axis Bank"],
                ["product_name", "Product Name", "text", "Flipkart Axis"],
                [
                  "linked_phone_number",
                  "Linked Phone",
                  "tel",
                  "10-digit mobile number",
                ],
                ["exp_month", "Expiry Month", "number", "MM"],
                ["exp_year", "Expiry Year", "number", "YYYY"],
                ["cvv", "CVV", "password", "3 digits"],
              ].map(([name, label, type, placeholder]) => (
                <label
                  key={name}
                  className='text-sm font-medium text-gray-700 dark:text-slate-300'
                >
                  {label}
                  <input
                    name={name}
                    type={type}
                    value={form[name]}
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
