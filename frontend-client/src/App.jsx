import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Main from './pages/Main';
import Login from './pages/Login';
import Signup from './pages/Signup';
import ShoppingCart from './pages/ShoppingCart';
import ProductDetail from './pages/ProductDetail';
import Profile from './pages/Profile';
import MyOrders from "./pages/MyOrders";
import Payment from './pages/Payment';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Main />} />
        <Route path="/login" element={<Login />} />
        <Route path="/signup" element={<Signup />} />
        <Route path="/shoppingcart" element={<ShoppingCart />} />
        <Route path="/products/:id" element={<ProductDetail />} />
        <Route path="/profile" element={<Profile />} />
        <Route path="/orders" element={<MyOrders />} />
        <Route path="/payment" element={<Payment />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;