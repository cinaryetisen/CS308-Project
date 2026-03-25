import { Link } from 'react-router-dom';

export default function Login() {
  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-100">
      <div className="w-full max-w-md p-8 space-y-6 bg-white rounded-lg shadow-md">
        <h2 className="text-3xl font-bold text-center text-gray-900">Log In</h2>
        
        <form className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700">Email Address</label>
            <input 
              type="email" 
              className="w-full px-4 py-2 mt-1 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500" 
              placeholder="Enter your email"
              required 
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-gray-700">Password</label>
            <input 
              type="password" 
              className="w-full px-4 py-2 mt-1 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500" 
              placeholder="Enter your password"
              required 
            />
          </div>
          
          <button 
            type="submit" 
            className="w-full px-4 py-2 font-bold text-white bg-blue-600 rounded hover:bg-blue-700 focus:outline-none"
          >
            Log In
          </button>
        </form>

        <div className="text-sm text-center text-gray-600">
          <p>
            Don't have an account? <Link to="/signup" className="text-blue-600 hover:underline">Sign up here</Link>
          </p>
          <p className="mt-2">
            <Link to="/" className="text-gray-500 hover:underline">Return to Main Page</Link>
          </p>
        </div>
      </div>
    </div>
  );
}