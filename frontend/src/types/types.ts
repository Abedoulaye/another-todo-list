export interface User {
    id: number;
    email: string;
    role: string;
}

export interface Note {
    id: number;
    userID: number;
    title: string;
    body: string;
}