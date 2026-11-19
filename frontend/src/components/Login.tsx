import { useState, type SubmitEvent } from "react";
import { useNavigate, Link} from "react-router";
import { useAuth } from "../context/AuthContext";

function Login() {
    const {login} = useAuth();
    const navifate = useNavigate();
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState("");
    const [loading, setLoaading] = useState(false);

    const handleSubmit = async (e: SubmitEvent) => {
        e.preventDefault();
        setError("");
        setLoaading(true);
    

    try {
        await login(email, password);
        // navigate("/");

    }
    catch (err){
        setError("Invalid email or password");
    }
    finally {
        setLoaading(false);
    }
}

}