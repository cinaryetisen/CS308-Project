import { Link, useNavigate } from 'react-router-dom';
import { useState } from 'react';
import { apiRequest } from '../api/client';

export default function Signup() {
    const navigate = useNavigate();

    const [formData, setFormData] = useState({
        fullName: "",
        email:    "",
        taxId:    "",
        address:  "",
        password: ""
    });

    const [loading, setLoading] = useState(false);
    const [error, setError]     = useState("");
    const [success, setSuccess] = useState(false);

    const handleChange = (e) => {
        setError("");
        setFormData({ ...formData, [e.target.name]: e.target.value });
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setLoading(true);
        setError("");

        try {
            await apiRequest("/api/signup", {
                method: "POST",
                body:   JSON.stringify({
                    name:         formData.fullName,
                    email:        formData.email,
                    tax_id:       formData.taxId,
                    home_address: formData.address,
                    password:     formData.password
                })
            }, false);

            setSuccess(true);
            setTimeout(() => navigate("/login"), 1500);

        } catch (err) {
            console.error(err);
            setError(err.message || "Signup failed. Please try again.");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-[#1a0f0a] flex items-center justify-center px-4 py-12">

            {/* Ambient glow */}
            <div className="absolute inset-0 pointer-events-none overflow-hidden">
                <div className="absolute top-1/3 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-[#8a47af]/10 rounded-full blur-[120px]" />
            </div>

            <div className="relative w-full max-w-lg">

                {/* Header */}
                <div className="text-center mb-8">
                    <Link to="/" className="text-3xl font-serif text-[#e7b4ff] hover:text-[#f5ded3] transition-colors">
                        The Vault
                    </Link>
                    <p className="mt-2 text-[#9a8c9b] text-sm tracking-widest uppercase">
                        Begin Your Journey
                    </p>
                </div>

                {/* Card */}
                <div className="bg-[#251912] border border-[#342720] rounded-lg p-8 shadow-[0_0_40px_rgba(138,71,175,0.08)]">

                    <h2 className="text-2xl font-serif text-[#f5ded3] mb-6">
                        Create an Account
                    </h2>

                    <form className="flex flex-col gap-4" onSubmit={handleSubmit}>

                        {/* Success */}
                        {success && (
                            <div className="px-4 py-3 text-sm text-[#131f00] bg-[#add461] rounded-lg font-medium">
                                ✓ Account created! Redirecting to login…
                            </div>
                        )}

                        {/* Error */}
                        {error && (
                            <div className="px-4 py-3 text-sm text-[#ffdad6] bg-[#93000a]/40 border border-[#93000a] rounded-lg">
                                {error}
                            </div>
                        )}

                        {/* Full Name */}
                        <div className="flex flex-col gap-1.5">
                            <label className="text-xs uppercase tracking-widest text-[#9a8c9b] font-semibold">
                                Full Name
                            </label>
                            <input
                                type="text"
                                name="fullName"
                                placeholder="Your full name"
                                value={formData.fullName}
                                onChange={handleChange}
                                className="w-full bg-[#1a0f0a] border border-[#342720] rounded px-4 py-2.5 text-sm text-[#f5ded3] placeholder-[#9a8c9b] focus:outline-none focus:border-[#e7b4ff] transition-colors"
                                required
                            />
                        </div>

                        {/* Email */}
                        <div className="flex flex-col gap-1.5">
                            <label className="text-xs uppercase tracking-widest text-[#9a8c9b] font-semibold">
                                Email Address
                            </label>
                            <input
                                type="text"
                                inputMode="email"
                                name="email"
                                placeholder="you@example.com"
                                value={formData.email}
                                onChange={handleChange}
                                className="w-full bg-[#1a0f0a] border border-[#342720] rounded px-4 py-2.5 text-sm text-[#f5ded3] placeholder-[#9a8c9b] focus:outline-none focus:border-[#e7b4ff] transition-colors"
                                required
                            />
                        </div>

                        {/* Tax ID & Address side note */}
                        <div className="border-t border-[#342720] pt-4 flex flex-col gap-4">
                            <p className="text-[10px] uppercase tracking-widest text-[#9a8c9b]/60">
                                Billing Information
                            </p>

                            {/* Tax ID */}
                            <div className="flex flex-col gap-1.5">
                                <label className="text-xs uppercase tracking-widest text-[#9a8c9b] font-semibold">
                                    Tax ID
                                </label>
                                <input
                                    type="text"
                                    name="taxId"
                                    placeholder="Your tax identification number"
                                    value={formData.taxId}
                                    onChange={handleChange}
                                    className="w-full bg-[#1a0f0a] border border-[#342720] rounded px-4 py-2.5 text-sm text-[#f5ded3] placeholder-[#9a8c9b] focus:outline-none focus:border-[#e7b4ff] transition-colors"
                                    required
                                />
                            </div>

                            {/* Home Address */}
                            <div className="flex flex-col gap-1.5">
                                <label className="text-xs uppercase tracking-widest text-[#9a8c9b] font-semibold">
                                    Home Address
                                </label>
                                <textarea
                                    name="address"
                                    placeholder="Your full address"
                                    value={formData.address}
                                    onChange={handleChange}
                                    rows={3}
                                    className="w-full bg-[#1a0f0a] border border-[#342720] rounded px-4 py-2.5 text-sm text-[#f5ded3] placeholder-[#9a8c9b] focus:outline-none focus:border-[#e7b4ff] transition-colors resize-none"
                                    required
                                />
                            </div>
                        </div>

                        {/* Password */}
                        <div className="border-t border-[#342720] pt-4 flex flex-col gap-1.5">
                            <label className="text-xs uppercase tracking-widest text-[#9a8c9b] font-semibold">
                                Password
                            </label>
                            <input
                                type="password"
                                name="password"
                                placeholder="••••••••"
                                value={formData.password}
                                onChange={handleChange}
                                className="w-full bg-[#1a0f0a] border border-[#342720] rounded px-4 py-2.5 text-sm text-[#f5ded3] placeholder-[#9a8c9b] focus:outline-none focus:border-[#e7b4ff] transition-colors"
                                required
                            />
                        </div>

                        {/* Submit */}
                        <button
                            type="submit"
                            disabled={loading || success}
                            className={`w-full py-3 rounded font-semibold transition-all duration-150 active:scale-95 mt-2 ${
                                loading || success
                                    ? "bg-[#342720] text-[#9a8c9b] cursor-not-allowed"
                                    : "bg-gradient-to-r from-[#e7b4ff] to-[#8a47af] text-[#300049] hover:brightness-110"
                            }`}
                        >
                            {loading ? "Creating your account…" : "Create Account"}
                        </button>

                    </form>

                    {/* Divider */}
                    <div className="border-t border-[#342720] my-6" />

                    {/* Links */}
                    <div className="flex flex-col gap-2 text-sm text-center text-[#9a8c9b]">
                        <p>
                            Already have an account?{" "}
                            <Link to="/login" className="text-[#e7b4ff] hover:underline font-medium">
                                Log in
                            </Link>
                        </p>
                        <Link to="/" className="text-[#9a8c9b]/60 hover:text-[#9a8c9b] transition-colors text-xs">
                            ← Return to the Vault
                        </Link>
                    </div>

                </div>
            </div>
        </div>
    );
}