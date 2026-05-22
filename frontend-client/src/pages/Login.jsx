import { Link, useNavigate } from 'react-router-dom';
import { useState } from 'react';

export default function Login() {
    const navigate = useNavigate();
    const API_URL  = import.meta.env.VITE_API_URL;

    const [formData, setFormData] = useState({ email: "", password: "" });
    const [loading, setLoading]   = useState(false);
    const [error, setError]       = useState("");
    const [success, setSuccess]   = useState(false);

    const handleChange = (e) => {
        setError("");
        setFormData({ ...formData, [e.target.name]: e.target.value });
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setLoading(true);
        setError("");

        try {
            const response = await fetch(`${API_URL}/api/login`, {
                method:  "POST",
                headers: { "Content-Type": "application/json" },
                body:    JSON.stringify({ email: formData.email, password: formData.password })
            });

            const data = await response.json();

            if (!response.ok) {
                setError(data.error || "Login failed. Please check your credentials.");
                setLoading(false);
                return;
            }

            if (data.token) localStorage.setItem("token", data.token);
            if (data.user)  localStorage.setItem("user", JSON.stringify(data.user));

            const guestCart = JSON.parse(localStorage.getItem("cart") || "[]");
            if (guestCart.length > 0) {
                try {
                    await fetch(`${API_URL}/api/cart/merge`, {
                        method:  "POST",
                        headers: {
                            "Content-Type":  "application/json",
                            "Authorization": `Bearer ${data.token}`
                        },
                        body: JSON.stringify(
                            guestCart.map((item) => ({ product_id: item.id, quantity: item.quantity }))
                        )
                    });
                } catch (mergeErr) {
                    console.error("Cart merge failed:", mergeErr);
                }
            }
            localStorage.removeItem("cart");

            setSuccess(true);

            try {
                const payload = JSON.parse(atob(data.token.split(".")[1]));
                setTimeout(() => {
                    if (payload.role === "product_manager") {
                        navigate("/pm/deliveries");
                    } else {
                        navigate("/");
                    }
                }, 1500);
            } catch {
                setTimeout(() => navigate("/"), 1500);
            }

        } catch (err) {
            console.error(err);
            setError("Server error. Please try again later.");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-[#1a0f0a] flex items-center justify-center px-4 py-12">

            {/* Ambient glow */}
            <div className="absolute inset-0 pointer-events-none overflow-hidden">
                <div className="absolute top-1/3 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[500px] h-[500px] bg-[#8a47af]/10 rounded-full blur-[120px]" />
            </div>

            <div className="relative w-full max-w-md">

                {/* Header */}
                <div className="text-center mb-8">
                    <Link to="/" className="text-3xl font-serif text-[#e7b4ff] hover:text-[#f5ded3] transition-colors">
                        The Vault
                    </Link>
                    <p className="mt-2 text-[#9a8c9b] text-sm tracking-widest uppercase">
                        Enter the Archive
                    </p>
                </div>

                {/* Card */}
                <div className="bg-[#251912] border border-[#342720] rounded-lg p-8 shadow-[0_0_40px_rgba(138,71,175,0.08)]">

                    <h2 className="text-2xl font-serif text-[#f5ded3] mb-6">
                        Welcome Back
                    </h2>

                    <form className="flex flex-col gap-4" onSubmit={handleSubmit}>

                        {/* Success */}
                        {success && (
                            <div className="px-4 py-3 text-sm text-[#131f00] bg-[#add461] rounded-lg font-medium">
                                ✓ Login successful! Redirecting…
                            </div>
                        )}

                        {/* Error */}
                        {error && (
                            <div className="px-4 py-3 text-sm text-[#ffdad6] bg-[#93000a]/40 border border-[#93000a] rounded-lg">
                                {error}
                            </div>
                        )}

                        {/* Email */}
                        <div className="flex flex-col gap-1.5">
                            <label className="text-xs uppercase tracking-widest text-[#9a8c9b] font-semibold">
                                Email Address
                            </label>
                            <input
                                name="email"
                                placeholder="you@example.com"
                                value={formData.email}
                                onChange={handleChange}
                                className="w-full bg-[#1a0f0a] border border-[#342720] rounded px-4 py-2.5 text-sm text-[#f5ded3] placeholder-[#9a8c9b] focus:outline-none focus:border-[#e7b4ff] transition-colors"
                                required
                            />
                        </div>

                        {/* Password */}
                        <div className="flex flex-col gap-1.5">
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
                            {loading ? "Entering the Vault…" : "Log In"}
                        </button>

                    </form>

                    {/* Divider */}
                    <div className="border-t border-[#342720] my-6" />

                    {/* Links */}
                    <div className="flex flex-col gap-2 text-sm text-center text-[#9a8c9b]">
                        <p>
                            No account yet?{" "}
                            <Link to="/signup" className="text-[#e7b4ff] hover:underline font-medium">
                                Create one
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