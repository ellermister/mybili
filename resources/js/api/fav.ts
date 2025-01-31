export async function getFavList() {
    const response = await fetch('/api/fav');
    return response.json();
}

export async function getFavDetail(id: number) {
    const response = await fetch(`/api/fav/${id}`);
    return response.json();
}
