import { Link, useNavigate } from 'react-router-dom';
import { useState } from 'react';
import { apiRequest } from '../api/client';

export default function Login() {
  const navigate = useNavigate();

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
      const data = await apiRequest("/api/login", {
        method: "POST",
        body: JSON.stringify({ email: formData.email, password: formData.password })
      }, false);

      if (data.token) localStorage.setItem("token", data.token);
      if (data.user)  localStorage.setItem("user", JSON.stringify(data.user));

      // Merge guest cart into backend, then clear localStorage cart
      const guestCart = JSON.parse(localStorage.getItem("cart") || "[]");
      if (guestCart.length > 0) {
        try {
          await apiRequest("/api/cart/merge", {
            method: "POST",
            body: JSON.stringify(
              guestCart.map((item) => ({
                product_id: item.id,
                quantity:   item.quantity
              }))
            )
          });
        } catch (mergeErr) {
          console.error("Cart merge failed:", mergeErr);
        }
      }
      localStorage.removeItem("cart");

      setSuccess(true);

      // Decode token to check role and redirect accordingly
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
      setError(err.message || "Server error. Please try again later.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-100">
      <div className="w-full max-w-md p-8 space-y-6 bg-white rounded-lg shadow-md">

        <h2 className="text-3xl font-bold text-center text-gray-900">Log In</h2>

        <form className="space-y-4" onSubmit={handleSubmit}>

          {success && (
            <div className="px-4 py-2 text-sm text-green-700 bg-green-100 border border-green-300 rounded-md">
              Login successful! Redirecting...
            </div>
          )}

          {error && (
            <div className="px-4 py-2 text-sm text-red-700 bg-red-100 border border-red-300 rounded-md">
              {error}
            </div>
          )}

          <div>
            <label className="block text-sm font-medium text-gray-700">Email Address</label>
            <input
              name="email"
              placeholder="Enter your email"
              value={formData.email}
              onChange={handleChange}
              className="w-full px-4 py-2 mt-1 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700">Password</label>
            <input
              type="password"
              name="password"
              placeholder="Enter your password"
              value={formData.password}
              onChange={handleChange}
              className="w-full px-4 py-2 mt-1 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              required
            />
          </div>

          <button
            type="submit"
            disabled={loading || success}
            className={`w-full px-4 py-2 font-bold text-white rounded focus:outline-none ${
              loading || success ? "bg-gray-400 cursor-not-allowed" : "bg-blue-600 hover:bg-blue-700"
            }`}
          >
            {loading ? "Logging in..." : "Log In"}
          </button>

        </form>

        <div className="text-sm text-center text-gray-600">
          <p>
            Don't have an account?{" "}
            <Link to="/signup" className="text-blue-600 hover:underline">Sign up here</Link>
          </p>
          <p className="mt-2">
            <Link to="/" className="text-gray-500 hover:underline">Return to Main Page</Link>
          </p>
        </div>

      </div>
    </div>
  );
}