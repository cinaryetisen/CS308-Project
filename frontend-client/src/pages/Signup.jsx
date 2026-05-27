import { Link, useNavigate } from 'react-router-dom';
import { useState } from 'react';
import { apiRequest } from '../api/client';

export default function Signup() {
  const navigate = useNavigate();

  // Form state
  const [formData, setFormData] = useState({
    fullName: "",
    email: "",
    taxId: "",
    address: "",
    password: ""
  });

  const [loading, setLoading] = useState(false);
  const [error, setError]     = useState("");

  // Handle input changes
  const handleChange = (e) => {
    setError("");
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  // Handle form submit
  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      await apiRequest("/api/signup", {
        method: "POST",
        body: JSON.stringify({
          name: formData.fullName,       // map frontend -> backend
          email: formData.email,
          tax_id: formData.taxId,
          home_address: formData.address,
          password: formData.password
        })
      }, false);

      navigate("/login");

    } catch (err) {
      console.error(err);
      setError(err.message || "Signup failed. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  return (
      <div className="flex items-center justify-center min-h-screen py-10 bg-gray-100">
        <div className="w-full max-w-lg p-8 space-y-6 bg-white rounded-lg shadow-md">

          <h2 className="text-3xl font-bold text-center text-gray-900">
            Create an Account
          </h2>

          <form className="space-y-4" onSubmit={handleSubmit}>

            {error && (
              <div className="px-4 py-2 text-sm text-red-700 bg-red-100 border border-red-300 rounded-md">
                {error}
              </div>
            )}

            {/* Full Name */}
            <input
                type="text"
                name="fullName"
                placeholder="Full Name"
                value={formData.fullName}
                onChange={handleChange}
                className="w-full px-4 py-2 border rounded-md focus:ring-2 focus:ring-blue-500"
                required
            />

            {/* Email */}
            <input
                type="text"
                inputMode="email"
                name="email"
                placeholder="Email Address"
                value={formData.email}
                onChange={handleChange}
                className="w-full px-4 py-2 border rounded-md focus:ring-2 focus:ring-blue-500"
                required
            />

            {/* Tax ID */}
            <input
                type="text"
                name="taxId"
                placeholder="Tax ID"
                value={formData.taxId}
                onChange={handleChange}
                className="w-full px-4 py-2 border rounded-md focus:ring-2 focus:ring-blue-500"
                required
            />

            {/* Home Address */}
            <textarea
                name="address"
                placeholder="Home Address"
                value={formData.address}
                onChange={handleChange}
                rows="3"
                className="w-full px-4 py-2 border rounded-md focus:ring-2 focus:ring-blue-500"
                required
            />

            {/* Password */}
            <input
                type="password"
                name="password"
                placeholder="Password"
                value={formData.password}
                onChange={handleChange}
                className="w-full px-4 py-2 border rounded-md focus:ring-2 focus:ring-blue-500"
                required
            />

            {/* Submit Button */}
            <button
                type="submit"
                disabled={loading}
                className={`w-full px-4 py-2 font-bold text-white rounded ${
                    loading
                        ? "bg-gray-400 cursor-not-allowed"
                        : "bg-green-600 hover:bg-green-700"
                }`}
            >
              {loading ? "Creating..." : "Sign Up"}
            </button>

          </form>

          {/* Links */}
          <div className="text-sm text-center text-gray-600">
            <p>
              Already have an account?{" "}
              <Link to="/login" className="text-blue-600 hover:underline">
                Log in
              </Link>
            </p>
            <p className="mt-2">
              <Link to="/" className="text-gray-500 hover:underline">
                Return to Main Page
              </Link>
            </p>
          </div>

        </div>
      </div>
  );
}
