import { useState, useEffect } from "react";

const API_BASE = import.meta.env.VITE_API_URL;

function PriceModal({ product, onClose, onDone }) {
    const [price, setPrice]       = useState(product.price?.toString() || "");
    const [discount, setDiscount] = useState(product.discount?.toString() || "0");
    const [saving, setSaving]     = useState(false);
    const [error, setError]       = useState(null);
    const [success, setSuccess]   = useState(null);

    const discountedPrice = parseFloat(price) > 0 && parseFloat(discount) >= 0
        ? parseFloat(price) * (1 - parseFloat(discount) / 100)
        : null;

    async function handleSavePrice() {
        const p = parseFloat(price);
        if (isNaN(p) || p <= 0) { setError("Price must be greater than 0."); return; }
        setSaving(true); setError(null); setSuccess(null);
        try {
            const token = localStorage.getItem("token");
            const res = await fetch(`${API_BASE}/api/admin/products/${product.id}/price`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
                body: JSON.stringify({ price: p }),
            });
            if (!res.ok) { const e = await res.json(); throw new Error(e.error || "Failed"); }
            setSuccess("Price updated.");
            onDone();
        } catch (e) { setError(e.message); }
        finally { setSaving(false); }
    }

    async function handleSaveDiscount() {
        const d = parseFloat(discount);
        if (isNaN(d) || d < 0 || d > 100) { setError("Discount must be between 0 and 100."); return; }
        setSaving(true); setError(null); setSuccess(null);
        try {
            const token = localStorage.getItem("token");
            const res = await fetch(`${API_BASE}/api/admin/products/${product.id}/discount`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
                body: JSON.stringify({ discount: d }),
            });
            if (!res.ok) { const e = await res.json(); throw new Error(e.error || "Failed"); }
            setSuccess(d > 0 ? "Discount set. Wishlist users will be notified." : "Discount removed.");
            onDone();
        } catch (e) { setError(e.message); }
        finally { setSaving(false); }
    }

    const inputClass = "w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500";
    const labelClass = "block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1";

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4">
            <div className="bg-white rounded-xl shadow-xl w-full max-w-md">
                <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
                    <div>
                        <h2 className="text-base font-bold text-gray-900">Pricing — {product.name}</h2>
                        <p className="text-xs text-gray-400 mt-0.5">{product.category}</p>
                    </div>
                    <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-xl leading-none">×</button>
                </div>
                <div className="px-6 py-5 flex flex-col gap-5">

                    {/* Price */}
                    <div>
                        <label className={labelClass}>Base Price ($) *</label>
                        <div className="flex gap-2">
                            <input
                                className={inputClass}
                                type="number"
                                min="0.01"
                                step="0.01"
                                value={price}
                                onChange={e => setPrice(e.target.value)}
                                placeholder="0.00"
                            />
                            <button
                                onClick={handleSavePrice}
                                disabled={saving}
                                className="shrink-0 bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold px-4 py-2 rounded-lg disabled:opacity-50 transition-colors"
                            >
                                {saving ? "…" : "Set"}
                            </button>
                        </div>
                    </div>

                    <div className="h-px bg-gray-100" />

                    {/* Discount */}
                    <div>
                        <label className={labelClass}>Discount (%) — set 0 to remove</label>
                        <div className="flex gap-2">
                            <input
                                className={inputClass}
                                type="number"
                                min="0"
                                max="100"
                                step="0.1"
                                value={discount}
                                onChange={e => setDiscount(e.target.value)}
                                placeholder="0"
                            />
                            <button
                                onClick={handleSaveDiscount}
                                disabled={saving}
                                className="shrink-0 bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold px-4 py-2 rounded-lg disabled:opacity-50 transition-colors"
                            >
                                {saving ? "…" : "Set"}
                            </button>
                        </div>
                    </div>

                    {/* Preview */}
                    {discountedPrice !== null && parseFloat(discount) > 0 && (
                        <div className="bg-blue-50 border border-blue-100 rounded-lg px-4 py-3 text-sm text-blue-800">
                            Final price after discount: <span className="font-bold">${discountedPrice.toFixed(2)}</span>
                            {" "}<span className="text-blue-400 line-through">${parseFloat(price).toFixed(2)}</span>
                        </div>
                    )}

                    {error   && <p className="text-sm text-red-600">{error}</p>}
                    {success && <p className="text-sm text-green-600">✓ {success}</p>}

                    <div className="flex justify-end pt-1 border-t border-gray-100">
                        <button onClick={onClose} className="px-4 py-2 text-sm border border-gray-300 rounded-lg text-gray-600 hover:bg-gray-50">
                            Close
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default function PricingManager() {
    const [products, setProducts] = useState([]);
    const [loading, setLoading]   = useState(true);
    const [error, setError]       = useState(null);
    const [search, setSearch]     = useState("");
    const [selected, setSelected] = useState(null);

    useEffect(() => { fetchProducts(); }, []);

    async function fetchProducts() {
        try {
            const res = await fetch(`${API_BASE}/api/products`);
            if (!res.ok) throw new Error("Failed to load products");
            setProducts(await res.json());
        } catch (e) {
            setError(e.message);
        } finally {
            setLoading(false);
        }
    }

    const filtered = products.filter(p =>
        p.name?.toLowerCase().includes(search.toLowerCase()) ||
        p.category?.toLowerCase().includes(search.toLowerCase())
    );

    if (loading) {
        return (
            <div className="flex items-center justify-center py-24">
                <div className="flex flex-col items-center gap-3">
                    <div className="w-8 h-8 border-4 border-gray-200 border-t-blue-600 rounded-full animate-spin" />
                    <p className="text-sm text-gray-500">Loading products…</p>
                </div>
            </div>
        );
    }

    return (
        <div className="max-w-4xl mx-auto">
            <div className="flex items-center justify-between mb-6">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">Pricing</h1>
                    <p className="text-sm text-gray-400 mt-0.5">Set prices and discounts for all products</p>
                </div>
            </div>

            <input
                type="text"
                placeholder="Search by name or category…"
                value={search}
                onChange={e => setSearch(e.target.value)}
                className="w-full border border-gray-300 rounded-lg px-4 py-2 text-sm mb-4 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />

            {error && (
                <div className="bg-red-50 border border-red-200 text-red-700 text-sm rounded-lg px-4 py-3 mb-4">{error}</div>
            )}

            <div className="bg-white border border-gray-200 rounded-xl overflow-hidden">
                <table className="w-full text-sm">
                    <thead>
                    <tr className="bg-gray-50 text-gray-400 text-xs uppercase tracking-wide border-b border-gray-100">
                        <th className="text-left px-5 py-3 font-semibold">Product</th>
                        <th className="text-left px-4 py-3 font-semibold">Category</th>
                        <th className="text-right px-4 py-3 font-semibold">Price</th>
                        <th className="text-center px-4 py-3 font-semibold">Discount</th>
                        <th className="text-right px-4 py-3 font-semibold">Final Price</th>
                        <th className="px-5 py-3"></th>
                    </tr>
                    </thead>
                    <tbody>
                    {filtered.length === 0 ? (
                        <tr>
                            <td colSpan="6" className="text-center text-gray-400 italic py-10">No products found.</td>
                        </tr>
                    ) : filtered.map(product => {
                        const finalPrice = product.discount > 0
                            ? product.price * (1 - product.discount / 100)
                            : null;
                        return (
                            <tr key={product.id} className="border-t border-gray-100 hover:bg-gray-50 transition-colors">
                                <td className="px-5 py-3">
                                    <div className="flex items-center gap-3">
                                        {product.image_url && (
                                            <img src={product.image_url} alt={product.name} className="w-9 h-9 rounded-lg object-cover border border-gray-100 shrink-0" />
                                        )}
                                        <p className="font-medium text-gray-900 line-clamp-1">{product.name}</p>
                                    </div>
                                </td>
                                <td className="px-4 py-3 text-gray-600">{product.category}</td>
                                <td className="px-4 py-3 text-gray-900 font-medium text-right">
                                    ${Number(product.price).toFixed(2)}
                                </td>
                                <td className="px-4 py-3 text-center">
                                    {product.discount > 0 ? (
                                        <span className="text-xs font-semibold px-2.5 py-1 rounded-full bg-red-100 text-red-700">
                                                -{product.discount}%
                                            </span>
                                    ) : (
                                        <span className="text-xs text-gray-300">—</span>
                                    )}
                                </td>
                                <td className="px-4 py-3 text-right">
                                    {finalPrice ? (
                                        <span className="font-semibold text-blue-600">${finalPrice.toFixed(2)}</span>
                                    ) : (
                                        <span className="text-gray-500">${Number(product.price).toFixed(2)}</span>
                                    )}
                                </td>
                                <td className="px-5 py-3 text-right">
                                    <button
                                        onClick={() => setSelected(product)}
                                        className="text-xs text-blue-600 border border-blue-200 hover:bg-blue-50 px-2.5 py-1 rounded-lg transition-colors"
                                    >
                                        Edit Pricing
                                    </button>
                                </td>
                            </tr>
                        );
                    })}
                    </tbody>
                </table>
            </div>

            {selected && (
                <PriceModal
                    product={selected}
                    onClose={() => setSelected(null)}
                    onDone={() => { setSelected(null); fetchProducts(); }}
                />
            )}
        </div>
    );
}