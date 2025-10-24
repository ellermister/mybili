import type { Video } from "./fav";
import type { Cover } from "./cover";

export interface Subscription {
    id: number;
    name: string;
    url: string;
    type: string;
    mid: string;
    description: string;
    total: number;
    status: number;
    last_check_at: number;
    videos: Video[] | null;
    cover_info: Cover|null;
}

export async function getSubscriptionList(): Promise<Subscription[]> {
    const response = await fetch('/api/subscription');
    return response.json();
}

export async function updateSubscription(id: number, data: Partial<Subscription>): Promise<Subscription> {
    const response = await fetch(`/api/subscription/${id}`, {
        method: 'PUT',
        body: JSON.stringify(data),
    });
    return response.json();
}

export async function removeSubscription(id: number): Promise<void> {
    await fetch(`/api/subscription/${id}`, {
        method: 'DELETE',
    });
}

export async function createNewSubscription(data: Partial<Subscription>): Promise<Subscription> {
    try{
        const response = await fetch('/api/subscription', {
            method: 'POST',
            body: JSON.stringify(data),
        });
        if(!response.ok){
            const error = await response.json();
            throw new Error(error.message);
        }
        return response.json();
    }catch(error){
        console.error(error);
        throw error;
    }

}

export async function getSubscriptionDetail(id: number): Promise<Subscription> {
    const response = await fetch(`/api/subscription/${id}`);
    return response.json();
}