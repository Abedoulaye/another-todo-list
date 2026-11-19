import { apiFetch } from "./api";
import type { Note } from "../types/types";

export async function getNotes(): Promise<Note[]> {
    return apiFetch("/notes")
}

export async function getNote(id: number): Promise<Note> {
    return apiFetch(`"/notes${id}"`)
}

export async function createProduct(note: Omit<Note, "id">,
): Promise<Note> {
    return apiFetch("notes", {method: "POST", body: JSON.stringify(note), });
}

export async function updateProduct(id: number, note: Omit<Note, "id">, ): Promise<Note> {
    return apiFetch(`/notes${id}`, {
        method: "PUT",
        body: JSON.stringify(note),
    });
}

export async function deleteNote(id: number): Promise<void> {
    return apiFetch(`/notes/${id}`, {
        method: "DELETE", 
    });
}