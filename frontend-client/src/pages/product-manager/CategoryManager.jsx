import { useState, useEffect } from "react";

const API_BASE = import.meta.env.VITE_API_URL;

export default function CategoryManager() {
    const [categories, setCategories] = useState([]);
    const [loading, setLoading]       = useState(true);
    const [error, setError]           = useState(null);
    const [newName, setNewName]       = useState("");
    const [creating, setCreating]     = useState(false);
    const [createError, setCreateError] = useState(null);

    useEffect(() => { fetchCategories(); }, []);

    async function fetchCategories() {
        setLoading(true);
        try {
            const res = await fetch(`${API_BASE}/api/categories`);
            if (!res.ok) throw new Error("Failed to load categories");
            setCategories(await res.json());
        } catch (e) {
            setError(e.message);
        } finally {
            setLoading(false);
        }
    }

    async function handleCreate() {
        if (!newName.trim()) { setCreateError("Category name cannot be empty."); return; }
        setCreating(true);
        setCreateError(null);
        try {
            const token = localStorage.getItem("token");
            const res = await fetch(`${API_BASE}/api/categories`, {
                method: "POST",
                headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
                body: JSON.stringify({ name: newName.trim() }),
            });
            if (!res.ok) { const e = await res.json(); throw new Error(e.error || "Failed to create"); }
            setNewName("");
            fetchCategories();
        } catch (e) {
            setCreateError(e.message);
        } finally {
            setCreating(false);
        }
    }

    async function handleDelete(category) {
        if (!window.confirm(`Delete category "${category.name}"? Products in this category will not be deleted.`)) return;
        try {
            const token = localStorage.getItem("token");
            const res = await fetch(`${API_BASE}/api/categories/${category.id}`, {
                method: "DELETE",
                headers: { Authorization: `Bearer ${token}` },
            });
            if (!res.ok) throw new Error("Failed to delete category");
            fetchCategories();
        } catch (e) {
            alert(e.message);
        }
    }

    if (loading) {
        return (
            <div className="flex items-center justify-center py-24">
                <div className="flex flex-col items-center gap-3">
                    <div className="w-8 h-8 border-4 border-gray-200 border-t-blue-600 rounded-full animate-spin" />
                    <p className="text-sm text-gray-500">Loading categories…</p>
                </div>
            </div>
        );
    }

    return (
        <div className="max-w-lg mx-auto">
            <div className="mb-6">
                <h1 className="text-2xl font-bold text-gray-900">Categories</h1>
                <p className="text-sm text-gray-400 mt-0.5">{categories.length} categories</p>
            </div>

            {/* Add new category */}
            <div className="bg-white border border-gray-200 rounded-xl p-5 mb-4">
                <h2 className="text-sm font-bold text-gray-700 mb-3">Add New Category</h2>
                <div className="flex gap-2">
                    <input
                        type="text"
                        placeholder="Category name…"
                        value={newName}
                        onChange={e => setNewName(e.target.value)}
                        onKeyDown={e => e.key === "Enter" && handleCreate()}
                        className="flex-1 border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                    <button
                        onClick={handleCreate}
                        disabled={creating}
                        className="bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold px-4 py-2 rounded-lg transition-colors disabled:opacity-50"
                    >
                        {creating ? "Adding…" : "Add"}
                    </button>
                </div>
                {createError && <p className="text-sm text-red-600 mt-2">{createError}</p>}
            </div>

            {error && (
                <div className="bg-red-50 border border-red-200 text-red-700 text-sm rounded-lg px-4 py-3 mb-4">{error}</div>
            )}

            {/* Categories list */}
            <div className="bg-white border border-gray-200 rounded-xl overflow-hidden">
                {categories.length === 0 ? (
                    <div className="text-center py-12">
                        <p className="text-3xl mb-2">🗂️</p>
                        <p className="text-gray-600 font-semibold">No categories yet</p>
                        <p className="text-sm text-gray-400 mt-1">Add your first category above.</p>
                    </div>
                ) : (
                    <ul className="divide-y divide-gray-100">
                        {categories.map(cat => (
                            <li key={cat.id || cat.name} className="flex items-center justify-between px-5 py-3 hover:bg-gray-50 transition-colors">
                                <span className="text-sm font-medium text-gray-800">{cat.name}</span>
                                <button
                                    onClick={() => handleDelete(cat)}
                                    className="text-xs text-red-600 border border-red-200 hover:bg-red-50 px-2.5 py-1 rounded-lg transition-colors"
                                >
                                    Delete
                                </button>
                            </li>
                        ))}
                    </ul>
                )}
            </div>
        </div>
    );
}