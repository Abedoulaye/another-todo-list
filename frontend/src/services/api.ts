const API_URL = import.meta.env.VITE_API_URL;

export async function apiFetch(endpoint: string, options: RequestInit = {}): Promise<any>{
    const response = await makeRequest(endpoint, options);
    
    // reminder 401 is unauthorized
    if (response.status === 401 && !endpoint.includes("/refresh")){
        const refreshed = await tryRefresh();
        if (refreshed){
            return apiFetch(endpoint, options);
        }
// if something gets down here it means refresh failed, localStorage.clear() logs out since our tokens are there
        localStorage.clear()
        window.location.href = "/login";
        throw new Error("Session expired");
    }

    if (!response.ok) {
        const error = await response.text();
        throw new Error(error || `Request failed with status ${response.status}`)
    }
// reminder 204 http.StatusNoContent
    if (response.status === 204) return null;

    return response.json();
}

async function makeRequest(endpoint: string, options: RequestInit): Promise<any> {
    const token = localStorage.getItem("token");
    const headers: Record<string, string> = {"Content-Type": "application/json", ...(options.headers as Record<string, string>), };

    if (token) headers["Authorization"] = `Bearer ${token}`;
    return fetch(`${API_URL}${endpoint}`, {...options, headers});
}

async function tryRefresh(): Promise<boolean> {
    const refreshToken = localStorage.getItem("refresh_token")
    if (!refreshToken) return false;

    try {
        const res = await fetch(`${API_URL}/refresh`, {
            method: "POST",
            headers: {"Content-Type": "apllication/json"},
            body: JSON.stringify({refresh_token: refreshToken})
        })
        if (!res.ok) return false;

        const data = await res.json();
        localStorage.setItem("token", data.access_token)
        return true;
    }
    catch {
        return false;
    }
}