declare module 'vue-virtual-scroller' {
    import { DefineComponent, Plugin } from 'vue';

    export interface RecycleScrollerProps {
        items: any[];
        itemSize: number | string;
        keyField?: string;
        buffer?: number;
        [key: string]: any;
    }

    export interface RecycleScrollerSlots {
        default(props: { item: any; index: number }): any;
    }

    export interface DynamicScrollerProps {
        items: any[];
        minItemSize?: number;
        keyField?: string;
        buffer?: number;
        [key: string]: any;
    }

    export interface DynamicScrollerSlots {
        default(props: { item: any; active: boolean; index: number }): any;
    }

    export interface DynamicScrollerItemProps {
        item: any;
        active: boolean;
        dataIndex?: number;
        [key: string]: any;
    }

    export const RecycleScroller: DefineComponent<RecycleScrollerProps, {}, {}, {}, {}, {}, {}, {}, string, {}, {}, RecycleScrollerSlots>;
    export const DynamicScroller: DefineComponent<DynamicScrollerProps, {}, {}, {}, {}, {}, {}, {}, string, {}, {}, DynamicScrollerSlots>;
    export const DynamicScrollerItem: DefineComponent<DynamicScrollerItemProps, {}, {}, {}, {}, {}, {}, {}, string, {}, {}, { default: () => any }>;
    
    const plugin: Plugin;
    export default plugin;
}

