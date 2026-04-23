import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";

const API_BASE = import.meta.env.VITE_API_URL;

function getInitials(name) {
    if (!name) return "?";
    return name
        .split(" ")
        .map((w) => w[0])
        .slice(0, 2)
        .join("")
        .toUpperCase();
}

export default function UserProfile() {
    const [user, setUser] = useState(null);
    const [form, setForm] = useState({ name: "", tax_id: "", home_address: "" });
    const [status, setStatus] = useState(null);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [editMode, setEditMode] = useState(false);
    const navigate = useNavigate();

    useEffect(() => {
        fetchProfile();
    }, []);

    async function fetchProfile() {
        setLoading(true);
        try {
            const token = localStorage.getItem("token");
            const res = await fetch(`${API_BASE}/api/users/me`, {
                headers: { Authorization: `Bearer ${token}` },
            });
            if (!res.ok) throw new Error("Failed to load profile");
            const data = await res.json();
            setUser(data);
            setForm({
                name: data.name || "",
                tax_id: data.tax_id || "",
                home_address: data.home_address || "",
            });
        } catch (err) {
            setStatus({ type: "error", message: err.message });
        } finally {
            setLoading(false);
        }
    }

    async function handleSave() {
        setSaving(true);
        setStatus(null);
        try {
            const token = localStorage.getItem("token");
            const res = await fetch(`${API_BASE}/api/users/me`, {
                method: "PATCH",
                headers: {
                    "Content-Type": "application/json",
                    Authorization: `Bearer ${token}`,
                },
                body: JSON.stringify({
                    name: form.name,
                    tax_id: form.tax_id,
                    home_address: form.home_address,
                }),
            });
            if (!res.ok) {
                const err = await res.json();
                throw new Error(err.error || "Update failed");
            }
            const updated = await res.json();
            setUser(updated);
            setForm({
                name: updated.name || "",
                tax_id: updated.tax_id || "",
                home_address: updated.home_address || "",
            });
            setStatus({ type: "success", message: "Profile updated successfully." });
            setEditMode(false);
        } catch (err) {
            setStatus({ type: "error", message: err.message });
        } finally {
            setSaving(false);
        }
    }

    function handleCancel() {
        setForm({
            name: user?.name || "",
            tax_id: user?.tax_id || "",
            home_address: user?.home_address || "",
        });
        setStatus(null);
        setEditMode(false);
    }

    if (loading) {
        return (
            <div className="min-h-screen bg-gray-100 flex items-center justify-center">
                <div className="flex flex-col items-center gap-3">
                    <div className="w-8 h-8 border-4 border-gray-200 border-t-blue-600 rounded-full animate-spin" />
                    <p className="text-sm text-gray-500">Loading profile…</p>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-gray-100 py-8 px-4">
            <div className="max-w-2xl mx-auto">

                {/* Back link */}
                <button
                    className="text-blue-600 text-sm mb-5 hover:underline flex items-center gap-1"
                    onClick={() => navigate("/")}
                >
                    ← Back to products
                </button>

                {/* Profile header card */}
                <div className="bg-white border border-gray-200 rounded-xl p-5 mb-4 flex items-center gap-4">
                    <div className="w-14 h-14 rounded-full bg-blue-600 text-white flex items-center justify-center text-xl font-semibold shrink-0">
                        {getInitials(user?.name)}
                    </div>
                    <div className="flex-1 min-w-0">
                        <h1 className="text-xl font-bold text-gray-900">{user?.name || "—"}</h1>
                        <p className="text-sm text-gray-500 mt-0.5">{user?.email || "—"}</p>
                    </div>
                    {!editMode && (
                        <button
                            className="shrink-0 bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold px-5 py-2 rounded-lg transition-colors"
                            onClick={() => setEditMode(true)}
                        >
                            Edit Profile
                        </button>
                    )}
                </div>

                {/* Status banner */}
                {status && (
                    <div
                        className={`rounded-lg px-4 py-2.5 text-sm font-medium mb-4 border ${
                            status.type === "success"
                                ? "bg-green-50 text-green-700 border-green-200"
                                : "bg-red-50 text-red-700 border-red-200"
                        }`}
                    >
                        {status.message}
                    </div>
                )}

                {/* Account card */}
                <div className="bg-white border border-gray-200 rounded-xl p-6 mb-4">
                    <h2 className="text-base font-bold text-gray-900 pb-3 mb-4 border-b border-gray-100">
                        Account
                    </h2>
                    <div className="flex flex-col gap-1">
                        <label className="text-xs font-semibold text-gray-400 uppercase tracking-wide">
                            Email
                        </label>
                        <p className="text-sm text-gray-900">{user?.email || "—"}</p>
                    </div>
                    <p className="text-xs text-gray-400 mt-3">Email address cannot be changed.</p>
                </div>

                {/* Personal details card */}
                <div className="bg-white border border-gray-200 rounded-xl p-6 mb-4">
                    <h2 className="text-base font-bold text-gray-900 pb-3 mb-4 border-b border-gray-100">
                        Personal Details
                    </h2>

                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">

                        {/* Full Name */}
                        <div className="flex flex-col gap-1">
                            <label className="text-xs font-semibold text-gray-400 uppercase tracking-wide">
                                Full Name
                            </label>
                            {editMode ? (
                                <input
                                    className="text-sm border border-gray-300 rounded-lg px-3 py-2 w-full focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                    value={form.name}
                                    onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                                    placeholder="Enter your name"
                                />
                            ) : (
                                <p className="text-sm text-gray-900">
                                    {user?.name || <span className="text-gray-300 italic">Not set</span>}
                                </p>
                            )}
                        </div>

                        {/* Tax ID */}
                        <div className="flex flex-col gap-1">
                            <label className="text-xs font-semibold text-gray-400 uppercase tracking-wide">
                                Tax ID
                            </label>
                            {editMode ? (
                                <input
                                    className="text-sm border border-gray-300 rounded-lg px-3 py-2 w-full focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                    value={form.tax_id}
                                    onChange={(e) => setForm((f) => ({ ...f, tax_id: e.target.value }))}
                                    placeholder="Enter your tax ID"
                                />
                            ) : (
                                <p className="text-sm text-gray-900">
                                    {user?.tax_id || <span className="text-gray-300 italic">Not set</span>}
                                </p>
                            )}
                        </div>

                        {/* Home Address */}
                        <div className="flex flex-col gap-1 sm:col-span-2">
                            <label className="text-xs font-semibold text-gray-400 uppercase tracking-wide">
                                Home Address
                            </label>
                            {editMode ? (
                                <textarea
                                    className="text-sm border border-gray-300 rounded-lg px-3 py-2 w-full focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-y"
                                    value={form.home_address}
                                    onChange={(e) => setForm((f) => ({ ...f, home_address: e.target.value }))}
                                    placeholder="Enter your address"
                                    rows={3}
                                />
                            ) : (
                                <p className="text-sm text-gray-900">
                                    {user?.home_address || <span className="text-gray-300 italic">Not set</span>}
                                </p>
                            )}
                        </div>
                    </div>

                    {/* Action buttons */}
                    {editMode && (
                        <div className="flex justify-end gap-2 mt-5 pt-5 border-t border-gray-100">
                            <button
                                className="px-5 py-2 text-sm font-medium text-gray-600 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors disabled:opacity-50"
                                onClick={handleCancel}
                                disabled={saving}
                            >
                                Cancel
                            </button>
                            <button
                                className="px-5 py-2 text-sm font-semibold text-white bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-60"
                                onClick={handleSave}
                                disabled={saving}
                            >
                                {saving ? "Saving…" : "Save Changes"}
                            </button>
                        </div>
                    )}
                </div>

            </div>
        </div>
    );
}