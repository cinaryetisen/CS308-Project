import { useState, useEffect } from "react";

const API_BASE = import.meta.env.VITE_API_URL;

const EMPTY_FORM = {
    name: "", model: "", serial_number: "", description: "",
    quantity: 0, category: "", distributor: "",
    warranty: "", image_url: "", tags: "",
};

function Modal({ title, onClose, children }) {
    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4">
            <div className="bg-white rounded-xl shadow-xl w-full max-w-xl max-h-[90vh] overflow-y-auto">
                <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
                    <h2 className="text-base font-bold text-gray-900">{title}</h2>
                    <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-xl leading-none">×</button>
                </div>
                <div className="px-6 py-5">{children}</div>
            </div>
        </div>
    );
}

function ProductForm({ initial, categories, onSubmit, onClose, loading }) {
    const [form, setForm] = useState(initial || EMPTY_FORM);
    const set = (k, v) => setForm(f => ({ ...f, [k]: v }));

    const inputClass = "w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500";
    const labelClass = "block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1";

    return (
        <div className="flex flex-col gap-4">
            <div className="grid grid-cols-2 gap-4">
                <div className="col-span-2">
                    <label className={labelClass}>Name *</label>
                    <input className={inputClass} value={form.name} onChange={e => set("name", e.target.value)} placeholder="Product name" />
                </div>
                <div>
                    <label className={labelClass}>Model *</label>
                    <input className={inputClass} value={form.model} onChange={e => set("model", e.target.value)} placeholder="Model number" />
                </div>
                <div>
                    <label className={labelClass}>Serial Number *</label>
                    <input className={inputClass} value={form.serial_number} onChange={e => set("serial_number", e.target.value)} placeholder="Serial number" />
                </div>
                <div className="col-span-2">
                    <label className={labelClass}>Description *</label>
                    <textarea className={`${inputClass} resize-none`} rows={3} value={form.description} onChange={e => set("description", e.target.value)} placeholder="Product description" />
                </div>
                <div>
                    <label className={labelClass}>Quantity</label>
                    <input className={inputClass} type="number" min="0" value={form.quantity} onChange={e => set("quantity", parseInt(e.target.value) || 0)} placeholder="0" />
                </div>
                <div>
                    <label className={labelClass}>Category *</label>
                    <select className={inputClass} value={form.category} onChange={e => set("category", e.target.value)}>
                        <option value="">Select category</option>
                        {categories.map(c => <option key={c.id || c.name} value={c.name}>{c.name}</option>)}
                    </select>
                </div>
                <div>
                    <label className={labelClass}>Distributor *</label>
                    <input className={inputClass} value={form.distributor} onChange={e => set("distributor", e.target.value)} placeholder="Distributor name" />
                </div>
                <div>
                    <label className={labelClass}>Warranty *</label>
                    <input className={inputClass} value={form.warranty} onChange={e => set("warranty", e.target.value)} placeholder="e.g. 2 years" />
                </div>
                <div>
                    <label className={labelClass}>Image URL</label>
                    <input className={inputClass} value={form.image_url} onChange={e => set("image_url", e.target.value)} placeholder="https://..." />
                </div>
                <div className="col-span-2">
                    <label className={labelClass}>Tags (comma separated)</label>
                    <input className={inputClass} value={form.tags} onChange={e => set("tags", e.target.value)} placeholder="sword, medieval, iron" />
                </div>
            </div>

            <div className="flex justify-end gap-2 pt-2 border-t border-gray-100">
                <button onClick={onClose} className="px-4 py-2 text-sm border border-gray-300 rounded-lg text-gray-600 hover:bg-gray-50">Cancel</button>
                <button
                    onClick={() => onSubmit({
                        ...form,
                        quantity: parseInt(form.quantity) || 0,
                        tags: typeof form.tags === "string"
                            ? form.tags.split(",").map(t => t.trim()).filter(Boolean)
                            : form.tags,
                    })}
                    disabled={loading}
                    className="px-4 py-2 text-sm bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-semibold disabled:opacity-50"
                >
                    {loading ? "Saving…" : "Save Product"}
                </button>
            </div>
        </div>
    );
}

function StockModal({ product, onClose, onDone }) {
    const [delta, setDelta] = useState("");
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState(null);

    async function handleUpdate() {
        const d = parseInt(delta);
        if (isNaN(d) || d === 0) { setError("Enter a non-zero number."); return; }
        setSaving(true);
        setError(null);
        try {
            const token = localStorage.getItem("token");
            const res = await fetch(`${API_BASE}/api/admin/products/${product.id}/stock`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
                body: JSON.stringify({ delta: d }),
            });
            if (!res.ok) { const e = await res.json(); throw new Error(e.error || "Failed"); }
            onDone();
        } catch (e) {
            setError(e.message);
        } finally {
            setSaving(false);
        }
    }

    return (
        <Modal title={`Update Stock — ${product.name}`} onClose={onClose}>
            <div className="flex flex-col gap-4">
                <div className="bg-gray-50 rounded-lg px-4 py-3 text-sm text-gray-700">
                    Current stock: <span className="font-bold text-gray-900">{product.quantity}</span>
                </div>
                <div>
                    <label className="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1">
                        Delta (positive to add, negative to remove)
                    </label>
                    <input
                        type="number"
                        className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                        value={delta}
                        onChange={e => setDelta(e.target.value)}
                        placeholder="e.g. 10 or -3"
                    />
                </div>
                {delta !== "" && !isNaN(parseInt(delta)) && (
                    <p className="text-xs text-gray-500">
                        New stock will be: <span className="font-bold text-gray-800">{product.quantity + parseInt(delta)}</span>
                    </p>
                )}
                {error && <p className="text-sm text-red-600">{error}</p>}
                <div className="flex justify-end gap-2 pt-2 border-t border-gray-100">
                    <button onClick={onClose} className="px-4 py-2 text-sm border border-gray-300 rounded-lg text-gray-600 hover:bg-gray-50">Cancel</button>
                    <button
                        onClick={handleUpdate}
                        disabled={saving}
                        className="px-4 py-2 text-sm bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-semibold disabled:opacity-50"
                    >
                        {saving ? "Saving…" : "Update Stock"}
                    </button>
                </div>
            </div>
        </Modal>
    );
}

export default function ProductManager() {
    const [products, setProducts]     = useState([]);
    const [categories, setCategories] = useState([]);
    const [loading, setLoading]       = useState(true);
    const [error, setError]           = useState(null);
    const [search, setSearch]         = useState("");

    const [showCreate, setShowCreate] = useState(false);
    const [editProduct, setEditProduct] = useState(null);
    const [stockProduct, setStockProduct] = useState(null);
    const [formLoading, setFormLoading]   = useState(false);
    const [formError, setFormError]       = useState(null);

    useEffect(() => { fetchAll(); }, []);

    async function fetchAll() {
        setLoading(true);
        try {
            const token = localStorage.getItem("token");
            const [pRes, cRes] = await Promise.all([
                // include_pending so the PM can manage products that are still
                // awaiting a sales-manager price (hidden from the storefront).
                fetch(`${API_BASE}/api/products?include_pending=true`),
                fetch(`${API_BASE}/api/categories`),
            ]);
            setProducts(pRes.ok ? await pRes.json() : []);
            setCategories(cRes.ok ? await cRes.json() : []);
        } catch (e) {
            setError(e.message);
        } finally {
            setLoading(false);
        }
    }

    async function handleCreate(form) {
        setFormLoading(true);
        setFormError(null);
        try {
            const token = localStorage.getItem("token");
            const res = await fetch(`${API_BASE}/api/admin/products`, {
                method: "POST",
                headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
                body: JSON.stringify(form),
            });
            if (!res.ok) { const e = await res.json(); throw new Error(e.error || "Failed to create"); }
            setShowCreate(false);
            fetchAll();
        } catch (e) {
            setFormError(e.message);
        } finally {
            setFormLoading(false);
        }
    }

    async function handleUpdate(form) {
        setFormLoading(true);
        setFormError(null);
        try {
            const token = localStorage.getItem("token");
            const res = await fetch(`${API_BASE}/api/admin/products/${editProduct.id}`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
                body: JSON.stringify(form),
            });
            if (!res.ok) { const e = await res.json(); throw new Error(e.error || "Failed to update"); }
            setEditProduct(null);
            fetchAll();
        } catch (e) {
            setFormError(e.message);
        } finally {
            setFormLoading(false);
        }
    }

    async function handleDelete(product) {
        if (!window.confirm(`Remove "${product.name}" from the catalog? This cannot be undone.`)) return;
        try {
            const token = localStorage.getItem("token");
            const res = await fetch(`${API_BASE}/api/admin/products/${product.id}`, {
                method: "DELETE",
                headers: { Authorization: `Bearer ${token}` },
            });
            if (!res.ok) throw new Error("Failed to delete");
            fetchAll();
        } catch (e) {
            alert(e.message);
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
        <div className="max-w-5xl mx-auto">
            {/* Header */}
            <div className="flex items-center justify-between mb-6">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">Products</h1>
                    <p className="text-sm text-gray-400 mt-0.5">{products.length} total products</p>
                </div>
                <button
                    onClick={() => { setShowCreate(true); setFormError(null); }}
                    className="bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold px-4 py-2 rounded-lg transition-colors"
                >
                    + Add Product
                </button>
            </div>

            {/* Search */}
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

            {/* Product table */}
            <div className="bg-white border border-gray-200 rounded-xl overflow-hidden">
                <table className="w-full text-sm">
                    <thead>
                    <tr className="bg-gray-50 text-gray-400 text-xs uppercase tracking-wide border-b border-gray-100">
                        <th className="text-left px-5 py-3 font-semibold">Product</th>
                        <th className="text-left px-4 py-3 font-semibold">Category</th>
                        <th className="text-right px-4 py-3 font-semibold">Price</th>
                        <th className="text-center px-4 py-3 font-semibold">Stock</th>
                        <th className="px-5 py-3"></th>
                    </tr>
                    </thead>
                    <tbody>
                    {filtered.length === 0 ? (
                        <tr>
                            <td colSpan="5" className="text-center text-gray-400 italic py-10">No products found.</td>
                        </tr>
                    ) : filtered.map(product => (
                        <tr key={product.id} className="border-t border-gray-100 hover:bg-gray-50 transition-colors">
                            <td className="px-5 py-3">
                                <div className="flex items-center gap-3">
                                    {product.image_url && (
                                        <img src={product.image_url} alt={product.name} className="w-9 h-9 rounded-lg object-cover border border-gray-100 shrink-0" />
                                    )}
                                    <div>
                                        <p className="font-medium text-gray-900 line-clamp-1">{product.name}</p>
                                        <p className="text-xs text-gray-400">{product.model}</p>
                                    </div>
                                </div>
                            </td>
                            <td className="px-4 py-3 text-gray-600">{product.category}</td>
                            <td className="px-4 py-3 text-gray-900 font-medium text-right">${Number(product.price).toFixed(2)}<span className="block text-xs text-gray-400 font-normal">Set by Sales</span></td>
                            <td className="px-4 py-3 text-center">
                                    <span className={`text-xs font-semibold px-2.5 py-1 rounded-full ${
                                        product.quantity === 0 ? "bg-red-100 text-red-700"
                                            : product.quantity <= 5 ? "bg-yellow-100 text-yellow-700"
                                                : "bg-green-100 text-green-700"
                                    }`}>
                                        {product.quantity}
                                    </span>
                            </td>
                            <td className="px-5 py-3">
                                <div className="flex items-center justify-end gap-2">
                                    <button
                                        onClick={() => setStockProduct(product)}
                                        className="text-xs text-blue-600 border border-blue-200 hover:bg-blue-50 px-2.5 py-1 rounded-lg transition-colors"
                                    >
                                        Stock
                                    </button>
                                    <button
                                        onClick={() => {
                                            setEditProduct(product);
                                            setFormError(null);
                                        }}
                                        className="text-xs text-gray-600 border border-gray-200 hover:bg-gray-50 px-2.5 py-1 rounded-lg transition-colors"
                                    >
                                        Edit
                                    </button>
                                    <button
                                        onClick={() => handleDelete(product)}
                                        className="text-xs text-red-600 border border-red-200 hover:bg-red-50 px-2.5 py-1 rounded-lg transition-colors"
                                    >
                                        Remove
                                    </button>
                                </div>
                            </td>
                        </tr>
                    ))}
                    </tbody>
                </table>
            </div>

            {/* Create modal */}
            {showCreate && (
                <Modal title="Add New Product" onClose={() => setShowCreate(false)}>
                    {formError && <p className="text-sm text-red-600 mb-4">{formError}</p>}
                    <ProductForm
                        categories={categories}
                        onSubmit={handleCreate}
                        onClose={() => setShowCreate(false)}
                        loading={formLoading}
                    />
                </Modal>
            )}

            {/* Edit modal */}
            {editProduct && (
                <Modal title="Edit Product" onClose={() => setEditProduct(null)}>
                    {formError && <p className="text-sm text-red-600 mb-4">{formError}</p>}
                    <ProductForm
                        initial={{
                            ...editProduct,
                            tags: Array.isArray(editProduct.tags) ? editProduct.tags.join(", ") : (editProduct.tags || ""),
                        }}
                        categories={categories}
                        onSubmit={handleUpdate}
                        onClose={() => setEditProduct(null)}
                        loading={formLoading}
                    />
                </Modal>
            )}

            {/* Stock modal */}
            {stockProduct && (
                <StockModal
                    product={stockProduct}
                    onClose={() => setStockProduct(null)}
                    onDone={() => { setStockProduct(null); fetchAll(); }}
                />
            )}
        </div>
    );
}