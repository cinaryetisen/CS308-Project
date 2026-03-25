import { Link } from 'react-router-dom';

export default function Main() {
  return (
    <div className="flex flex-col items-center justify-center h-screen bg-gray-50">
      <h1 className="text-4xl font-bold mb-4">Main Page</h1>
      <div className="space-x-4">
        <Link to="/login" className="text-blue-600 hover:underline">Go to Login</Link>
        <Link to="/signup" className="text-blue-600 hover:underline">Go to Sign Up</Link>
      </div>
    </div>
  );
}