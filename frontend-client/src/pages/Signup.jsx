import { Link, useNavigate } from 'react-router-dom';
import { useState } from 'react';

export default function Signup() {
  const navigate = useNavigate();

  // API URL from .env
  const API_URL = import.meta.env.VITE_API_URL;

  // Form state
  const [formData, setFormData] = useState({
    fullName: "",
    email: "",
    taxId: "",
    address: "",
    password: ""
  });

  const [loading, setLoading] = useState(false);

  // Handle input changes
  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value
    });
  };

  // Handle form submit
  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);

    try {
      const response = await fetch(`${API_URL}/api/signup`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          name: formData.fullName,       // map frontend -> backend
          email: formData.email,
          tax_id: formData.taxId,
          home_address: formData.address,
          password: formData.password
        })
      });

      const data = await response.json();

      if (!response.ok) {
        alert(data.error || "Signup failed");
        setLoading(false);
        return;
      }

      alert("Account created successfully!");
      navigate("/login");

    } catch (error) {
      console.error(error);
      alert("Server error");
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
                type="email"
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