import { useState, useEffect } from "react";
import { useNavigate, Link } from "react-router-dom";

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
    const [user, setUser]       = useState(null);
    const [form, setForm]       = useState({ name: "", tax_id: "", home_address: "" });
    const [status, setStatus]   = useState(null);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving]   = useState(false);
    const [editMode, setEditMode] = useState(false);
    const navigate = useNavigate();

    useEffect(() => { fetchProfile(); }, []);

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
            setForm({ name: data.name || "", tax_id: data.tax_id || "", home_address: data.home_address || "" });
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
                headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
                body: JSON.stringify({ name: form.name, tax_id: form.tax_id, home_address: form.home_address }),
            });
            if (!res.ok) {
                const err = await res.json();
                throw new Error(err.error || "Update failed");
            }
            const updated = await res.json();
            setUser(updated);
            setForm({ name: updated.name || "", tax_id: updated.tax_id || "", home_address: updated.home_address || "" });
            setStatus({ type: "success", message: "Profile updated successfully." });
            setEditMode(false);
        } catch (err) {
            setStatus({ type: "error", message: err.message });
        } finally {
            setSaving(false);
        }
    }

    function handleCancel() {
        setForm({ name: user?.name || "", tax_id: user?.tax_id || "", home_address: user?.home_address || "" });
        setStatus(null);
        setEditMode(false);
    }

    if (loading) {
        return (
            <div className="min-h-screen bg-[#1a0f0a] flex items-center justify-center">
                <p className="text-[#9a8c9b] tracking-widest animate-pulse">Summoning profile…</p>
            </div>
        );
    }

    const inputClass =
        "w-full bg-[#1a0f0a] border border-[#342720] px-4 py-2.5 rounded text-sm text-[#f5ded3] placeholder-[#9a8c9b] focus:outline-none focus:border-[#e7b4ff] transition-colors";

    const labelClass = "text-[10px] uppercase tracking-widest font-semibold text-[#9a8c9b]";

    return (
        <div className="min-h-screen bg-[#1a0f0a] flex flex-col">
            <main className="flex-1 max-w-2xl mx-auto w-full px-4 sm:px-6 py-8 sm:py-10 flex flex-col gap-5">

                {/* Breadcrumb */}
                <div className="text-sm text-[#9a8c9b] flex gap-2 items-center">
                    <Link to="/" className="hover:text-[#e7b4ff] transition-colors">The Vault</Link>
                    <span className="text-[#342720]">/</span>
                    <span className="text-[#f5ded3] font-medium">My Profile</span>
                </div>

                {/* Header card */}
                <div className="bg-[#251912] border border-[#342720] rounded-lg p-5 flex items-center gap-4 shadow-[0_0_30px_rgba(138,71,175,0.06)]">
                    <div className="w-14 h-14 rounded-full bg-[#8a47af]/20 border border-[#8a47af]/50 flex items-center justify-center text-xl font-bold text-[#e7b4ff] shrink-0">
                        {getInitials(user?.name)}
                    </div>
                    <div className="flex-1 min-w-0">
                        <h1 className="text-xl font-serif text-[#f5ded3] truncate">{user?.name || "—"}</h1>
                        <p className="text-sm text-[#9a8c9b] mt-0.5 truncate">{user?.email || "—"}</p>
                    </div>
                    {!editMode && (
                        <button
                            onClick={() => setEditMode(true)}
                            className="shrink-0 px-5 py-2 text-sm font-semibold rounded bg-gradient-to-r from-[#e7b4ff] to-[#8a47af] text-[#300049] hover:brightness-110 active:scale-95 transition-all"
                        >
                            Edit Profile
                        </button>
                    )}
                </div>

                {/* Status banner */}
                {status && (
                    <div className={`rounded-lg px-4 py-2.5 text-sm font-medium border ${
                        status.type === "success"
                            ? "bg-[#add461]/10 text-[#add461] border-[#add461]/30"
                            : "bg-[#93000a]/20 text-[#ffdad6] border-[#93000a]/50"
                    }`}>
                        {status.type === "success" ? "✓ " : "✕ "}{status.message}
                    </div>
                )}

                {/* Account card */}
                <div className="bg-[#251912] border border-[#342720] rounded-lg p-6">
                    <h2 className="text-base font-serif text-[#f5ded3] pb-3 mb-4 border-b border-[#342720]">
                        Account
                    </h2>
                    <div className="flex flex-col gap-1">
                        <span className={labelClass}>Email</span>
                        <p className="text-sm text-[#f5ded3]">{user?.email || "—"}</p>
                    </div>
                    <p className="text-xs text-[#9a8c9b]/60 mt-3">Email address cannot be changed.</p>
                </div>

                {/* Personal details card */}
                <div className="bg-[#251912] border border-[#342720] rounded-lg p-6">
                    <h2 className="text-base font-serif text-[#f5ded3] pb-3 mb-5 border-b border-[#342720]">
                        Personal Details
                    </h2>

                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-5">

                        {/* Full Name */}
                        <div className="flex flex-col gap-1.5">
                            <span className={labelClass}>Full Name</span>
                            {editMode ? (
                                <input
                                    className={inputClass}
                                    value={form.name}
                                    onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                                    placeholder="Enter your name"
                                />
                            ) : (
                                <p className="text-sm text-[#f5ded3]">
                                    {user?.name || <span className="text-[#9a8c9b]/50 italic">Not set</span>}
                                </p>
                            )}
                        </div>

                        {/* Tax ID */}
                        <div className="flex flex-col gap-1.5">
                            <span className={labelClass}>Tax ID</span>
                            {editMode ? (
                                <input
                                    className={inputClass}
                                    value={form.tax_id}
                                    onChange={(e) => setForm((f) => ({ ...f, tax_id: e.target.value }))}
                                    placeholder="Enter your tax ID"
                                />
                            ) : (
                                <p className="text-sm text-[#f5ded3]">
                                    {user?.tax_id || <span className="text-[#9a8c9b]/50 italic">Not set</span>}
                                </p>
                            )}
                        </div>

                        {/* Home Address */}
                        <div className="flex flex-col gap-1.5 sm:col-span-2">
                            <span className={labelClass}>Home Address</span>
                            {editMode ? (
                                <textarea
                                    className={`${inputClass} resize-y`}
                                    value={form.home_address}
                                    onChange={(e) => setForm((f) => ({ ...f, home_address: e.target.value }))}
                                    placeholder="Enter your address"
                                    rows={3}
                                />
                            ) : (
                                <p className="text-sm text-[#f5ded3] whitespace-pre-line">
                                    {user?.home_address || <span className="text-[#9a8c9b]/50 italic">Not set</span>}
                                </p>
                            )}
                        </div>
                    </div>

                    {editMode && (
                        <div className="flex justify-end gap-3 mt-6 pt-5 border-t border-[#342720]">
                            <button
                                onClick={handleCancel}
                                disabled={saving}
                                className="px-5 py-2 text-sm font-medium rounded border border-[#342720] text-[#9a8c9b] hover:border-[#8a47af] hover:text-[#e7b4ff] transition-all disabled:opacity-50"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleSave}
                                disabled={saving}
                                className="px-6 py-2 text-sm font-semibold rounded bg-gradient-to-r from-[#e7b4ff] to-[#8a47af] text-[#300049] hover:brightness-110 active:scale-95 transition-all disabled:opacity-60 disabled:cursor-not-allowed"
                            >
                                {saving ? "Saving…" : "Save Changes"}
                            </button>
                        </div>
                    )}
                </div>

                <Link
                    to="/"
                    className="text-center text-sm text-[#9a8c9b] hover:text-[#e7b4ff] transition-colors"
                >
                    ← Return to the Vault
                </Link>

            </main>
        </div>
    );
}
