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

const currentYear = new Date().getFullYear();
const maxExpiryYear = currentYear + 20;

const cardNumberPattern =
  /^(?:4\d{12}(?:\d{3})?|(?:5[1-5]\d{2}|222[1-9]|22[3-9]\d|2[3-6]\d{2}|27[01]\d|2720)\d{12}|6(?:011|5\d{2})\d{12}|(?:60|65|81|82|508)\d{13})$/;
const phonePattern = /^[6-9]\d{9}$/;
const expiryMonthPattern = /^(0[1-9]|1[0-2])$/;
const expiryYearPattern = /^\d{4}$/;
const integerFieldLimits = {
  pan: 16,
  linked_phone_number: 10,
  exp_month: 2,
  exp_year: 4,
  cvv: 3,
};

function validateCardForm(card) {
  const pan = String(card.pan ?? "").trim();
  const cardholderName = String(card.cardholder_name ?? "").trim();
  const bankName = String(card.bank_name ?? "").trim();
  const productName = String(card.product_name ?? "").trim();
  const linkedPhoneNumber = String(card.linked_phone_number ?? "").trim();
  const expiryMonth = String(card.exp_month ?? "");
  const expiryYear = String(card.exp_year ?? "");
  const expMonth = Number(expiryMonth);
  const expYear = Number(expiryYear);
  const cvv = String(card.cvv ?? "").trim();

  if (!cardNumberPattern.test(pan)) {
    return "Enter a valid 13- or 16-digit credit card number.";
  }
  if (!cardholderName || cardholderName.length > 100) {
    return "Cardholder name is required and must be 100 characters or fewer.";
  }
  if (!bankName || bankName.length > 100) {
    return "Bank name is required and must be 100 characters or fewer.";
  }
  if (productName.length > 100) {
    return "Product name must be 100 characters or fewer.";
  }
  if (!phonePattern.test(linkedPhoneNumber)) {
    return "Enter a valid 10-digit mobile number starting with 6-9.";
  }
  if (!expiryMonthPattern.test(expiryMonth)) {
    return "Expiry month must be exactly two digits, from 01 to 12.";
  }
  if (
    !expiryYearPattern.test(expiryYear) ||
    expYear < currentYear ||
    expYear > maxExpiryYear
  ) {
    return `Expiry year must be between ${currentYear} and ${maxExpiryYear}.`;
  }
  if (expYear === currentYear && expMonth < new Date().getMonth() + 1) {
    return "The card expiry date must be in the future.";
  }
  if (!/^\d{3}$/.test(cvv)) {
    return "CVV must be exactly 3 digits.";
  }
  return "";
}

function getCardFormValues(card = emptyCard) {
  return {
    pan: String(card.pan ?? ""),
    cardholder_name: String(card.cardholder_name ?? ""),
    bank_name: String(card.bank_name ?? ""),
    product_name: String(card.product_name ?? ""),
    payment_method_type: String(card.payment_method_type ?? "Credit"),
    card_brand: String(card.card_brand ?? "Visa"),
    linked_phone_number: String(card.linked_phone_number ?? ""),
    exp_month:
      card.exp_month === "" || card.exp_month == null
        ? ""
        : String(card.exp_month).padStart(2, "0"),
    exp_year: String(card.exp_year ?? ""),
    cvv: String(card.cvv ?? ""),
  };
}

function MyCards() {
  const [cards, setCards] = useState([]);
  const [form, setForm] = useState(emptyCard);
  const [originalForm, setOriginalForm] = useState(null);
  const [editingId, setEditingId] = useState(null);
  const [formOpen, setFormOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const showToast = useToast();

  const loadCards = async () => {
    setLoading(true);
    try {
      const data = await apiRequest("/api/cards/mine");
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
  const updateField = (event) => {
    const { name, value } = event.target;
    const limit = integerFieldLimits[name];
    const nextValue = limit ? value.replace(/\D/g, "").slice(0, limit) : value;

    setForm((current) => ({
      ...current,
      [name]: nextValue,
    }));
  };
  const closeForm = () => {
    setFormOpen(false);
    setEditingId(null);
    setOriginalForm(null);
    setError("");
    setForm(emptyCard);
  };
  const openCreate = () => {
    setForm({ ...emptyCard });
    setEditingId(null);
    setOriginalForm(null);
    setError("");
    setFormOpen(true);
  };
  const openEdit = (card) => {
    const cardForm = getCardFormValues(card);
    setEditingId(card.id);
    setForm(cardForm);
    setOriginalForm(cardForm);
    setError("");
    setFormOpen(true);
  };

  const saveDisabled =
    saving ||
    (editingId !== null &&
      JSON.stringify(form) === JSON.stringify(originalForm));

  const submit = async (event) => {
    event.preventDefault();
    const validationError = validateCardForm(form);
    if (validationError) {
      setError(validationError);
      showToast(validationError, "error");
      return;
    }
    setSaving(true);
    setError("");
    try {
      await apiRequest(editingId ? `/api/cards/${editingId}` : "/api/cards", {
        method: editingId ? "PUT" : "POST",
        body: JSON.stringify({
          ...form,
          exp_month: Number(form.exp_month),
          exp_year: Number(form.exp_year),
          cvv: Number(form.cvv),
        }),
      });
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
      await apiRequest(`/api/cards/${card.id}`, { method: "DELETE" });
      setCards((current) => current.filter((item) => item.id !== card.id));
      showToast("Card deleted successfully.", "success");
    } catch (requestError) {
      setError(requestError.message);
      showToast(requestError.message, "error");
    }
  };

  const fields = [
    {
      name: "pan",
      label: "Credit Card Number",
      type: "text",
      placeholder: "13 or 16 digits",
      maxLength: 16,
      inputMode: "numeric",
      pattern: "[0-9]{13}|[0-9]{16}",
      required: true,
    },
    {
      name: "cardholder_name",
      label: "Cardholder Name",
      type: "text",
      placeholder: "Name on card",
      maxLength: 100,
      required: true,
    },
    {
      name: "bank_name",
      label: "Bank",
      type: "text",
      placeholder: "Axis Bank",
      maxLength: 100,
      required: true,
    },
    {
      name: "product_name",
      label: "Product Name",
      type: "text",
      placeholder: "Flipkart Axis (optional)",
      maxLength: 100,
      required: false,
    },
    {
      name: "linked_phone_number",
      label: "Linked Phone",
      type: "tel",
      placeholder: "10-digit mobile number",
      maxLength: 10,
      inputMode: "numeric",
      pattern: "[6-9][0-9]{9}",
      required: true,
    },
    {
      name: "exp_month",
      label: "Expiry Month",
      type: "text",
      placeholder: "MM (01-12)",
      maxLength: 2,
      inputMode: "numeric",
      pattern: "(0[1-9]|1[0-2])",
      required: true,
    },
    {
      name: "exp_year",
      label: "Expiry Year",
      type: "text",
      placeholder: `YYYY (${currentYear}-${maxExpiryYear})`,
      maxLength: 4,
      inputMode: "numeric",
      pattern: "[0-9]{4}",
      required: true,
    },
    {
      name: "cvv",
      label: "CVV",
      type: "password",
      placeholder: "3 digits",
      maxLength: 3,
      inputMode: "numeric",
      pattern: "[0-9]{3}",
      required: true,
    },
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
      {error && !formOpen && (
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
            noValidate
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
            {error && (
              <p className='mb-4 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700 dark:bg-red-950 dark:text-red-300'>
                {error}
              </p>
            )}
            <div className='grid grid-cols-1 gap-4 sm:grid-cols-2'>
              {fields.map((field) => (
                <label
                  key={field.name}
                  className='text-sm font-medium text-gray-700 dark:text-slate-300'
                >
                  {field.label}
                  <input
                    name={field.name}
                    type={field.type}
                    value={form[field.name] || ""}
                    onChange={updateField}
                    placeholder={field.placeholder}
                    maxLength={field.maxLength}
                    min={field.min}
                    max={field.max}
                    inputMode={field.inputMode}
                    pattern={field.pattern}
                    required={field.required}
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
            <button
              type='submit'
              disabled={saveDisabled}
              className='btn-primary save-card-button mt-6 w-full'
            >
              {saving ? "Saving..." : "Save Card"}
            </button>
          </form>
        </div>
      )}
    </section>
  );
}

export default MyCards;
