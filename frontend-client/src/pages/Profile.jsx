import { useState, useEffect } from "react";
import { useNavigate, Link } from "react-router-dom";
import { apiRequest } from "../api/client";

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
            const data = await apiRequest("/api/users/me");
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
            const updated = await apiRequest("/api/users/me", {
                method: "PATCH",
                body: JSON.stringify({
                    name: form.name,
                    tax_id: form.tax_id,
                    home_address: form.home_address,
                }),
            });
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
            <div className="min-h-screen bg-[var(--bg)] flex items-center justify-center">
                <p className="text-[var(--muted)] tracking-widest animate-pulse">Summoning profile…</p>
            </div>
        );
    }

    const inputClass =
        "w-full bg-[var(--bg)] border border-[var(--border)] px-4 py-2.5 rounded text-sm text-[var(--text)] placeholder-[var(--muted)] focus:outline-none focus:border-[var(--accent)] transition-colors";

    const labelClass = "text-[10px] uppercase tracking-widest font-semibold text-[var(--muted)]";

    return (
        <div className="min-h-screen bg-[var(--bg)] flex flex-col">
            <main className="flex-1 max-w-2xl mx-auto w-full px-4 sm:px-6 py-8 sm:py-10 flex flex-col gap-5">

                {/* Breadcrumb */}
                <div className="text-sm text-[var(--muted)] flex gap-2 items-center">
                    <Link to="/" className="hover:text-[var(--accent)] transition-colors">The Vault</Link>
                    <span className="text-[var(--border)]">/</span>
                    <span className="text-[var(--text)] font-medium">My Profile</span>
                </div>

                {/* Header card */}
                <div className="bg-[var(--surface)] border border-[var(--border)] rounded-lg p-5 flex items-center gap-4 shadow-[0_0_30px_rgba(138,71,175,0.06)]">
                    <div className="w-14 h-14 rounded-full bg-[#8a47af]/20 border border-[#8a47af]/50 flex items-center justify-center text-xl font-bold text-[var(--accent)] shrink-0">
                        {getInitials(user?.name)}
                    </div>
                    <div className="flex-1 min-w-0">
                        <h1 className="text-xl font-serif text-[var(--text)] truncate">{user?.name || "—"}</h1>
                        <p className="text-sm text-[var(--muted)] mt-0.5 truncate">{user?.email || "—"}</p>
                    </div>
                    {!editMode && (
                        <button
                            onClick={() => setEditMode(true)}
                            className="shrink-0 px-5 py-2 text-sm font-semibold rounded bg-gradient-to-r from-[var(--btn-from)] to-[var(--btn-to)] text-[var(--on-accent)] hover:brightness-110 active:scale-95 transition-all"
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
                <div className="bg-[var(--surface)] border border-[var(--border)] rounded-lg p-6">
                    <h2 className="text-base font-serif text-[var(--text)] pb-3 mb-4 border-b border-[var(--border)]">
                        Account
                    </h2>
                    <div className="flex flex-col gap-1">
                        <span className={labelClass}>Email</span>
                        <p className="text-sm text-[var(--text)]">{user?.email || "—"}</p>
                    </div>
                    <p className="text-xs text-[var(--muted)]/60 mt-3">Email address cannot be changed.</p>
                </div>

                {/* Personal details card */}
                <div className="bg-[var(--surface)] border border-[var(--border)] rounded-lg p-6">
                    <h2 className="text-base font-serif text-[var(--text)] pb-3 mb-5 border-b border-[var(--border)]">
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
                                <p className="text-sm text-[var(--text)]">
                                    {user?.name || <span className="text-[var(--muted)]/50 italic">Not set</span>}
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
                                <p className="text-sm text-[var(--text)]">
                                    {user?.tax_id || <span className="text-[var(--muted)]/50 italic">Not set</span>}
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
                                <p className="text-sm text-[var(--text)] whitespace-pre-line">
                                    {user?.home_address || <span className="text-[var(--muted)]/50 italic">Not set</span>}
                                </p>
                            )}
                        </div>
                    </div>

                    {editMode && (
                        <div className="flex justify-end gap-3 mt-6 pt-5 border-t border-[var(--border)]">
                            <button
                                onClick={handleCancel}
                                disabled={saving}
                                className="px-5 py-2 text-sm font-medium rounded border border-[var(--border)] text-[var(--muted)] hover:border-[var(--accent-dim)] hover:text-[var(--accent)] transition-all disabled:opacity-50"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleSave}
                                disabled={saving}
                                className="px-6 py-2 text-sm font-semibold rounded bg-gradient-to-r from-[var(--btn-from)] to-[var(--btn-to)] text-[var(--on-accent)] hover:brightness-110 active:scale-95 transition-all disabled:opacity-60 disabled:cursor-not-allowed"
                            >
                                {saving ? "Saving…" : "Save Changes"}
                            </button>
                        </div>
                    )}
                </div>

                <Link
                    to="/"
                    className="text-center text-sm text-[var(--muted)] hover:text-[var(--accent)] transition-colors"
                >
                    ← Return to the Vault
                </Link>

            </main>
        </div>
    );
}
