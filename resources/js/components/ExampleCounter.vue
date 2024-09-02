<script>
export default {
    data() {
        return {
            count: 0,
        };
    },
};
</script>

<template>

    <div class="container p-4">

        <div class="row">
        <div class="offset-sm-4 col-sm-4">
            <button class="btn btn-primary btn-lg col-12 mt-5" @click="count++">
                <i class="bi bi-plus-circle"></i>
                You clicked me {{ count }} times.
            </button>
        </div>
    </div>
    </div>
   
</template>