import { Link } from 'react-router-dom';

export default function Signup() {
  return (
    <div className="flex items-center justify-center min-h-screen py-10 bg-gray-100">
      <div className="w-full max-w-lg p-8 space-y-6 bg-white rounded-lg shadow-md">
        <h2 className="text-3xl font-bold text-center text-gray-900">Create an Account</h2>
        
        <form className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700">Full Name</label>
            <input 
              type="text" 
              className="w-full px-4 py-2 mt-1 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500" 
              required 
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700">Email Address</label>
            <input 
              type="email" 
              className="w-full px-4 py-2 mt-1 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500" 
              required 
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700">Tax ID</label>
            <input 
              type="text" 
              className="w-full px-4 py-2 mt-1 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500" 
              required 
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700">Home Address</label>
            <textarea 
              className="w-full px-4 py-2 mt-1 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500" 
              rows="3" 
              required
            ></textarea>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700">Password</label>
            <input 
              type="password" 
              className="w-full px-4 py-2 mt-1 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500" 
              required 
            />
          </div>

          <button 
            type="submit" 
            className="w-full px-4 py-2 font-bold text-white bg-green-600 rounded hover:bg-green-700 focus:outline-none"
          >
            Sign Up
          </button>
        </form>

        <div className="text-sm text-center text-gray-600">
          <p>
            Already have an account? <Link to="/login" className="text-blue-600 hover:underline">Log in</Link>
          </p>
          <p className="mt-2">
            <Link to="/" className="text-gray-500 hover:underline">Return to Main Page</Link>
          </p>
        </div>
      </div>
    </div>
  );
}