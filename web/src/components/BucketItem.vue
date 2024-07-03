<script>
export default {
    computed: {
        currentPath() {
            return this.$route.path;
        }
    },
    props: {
        bucket: {
            type: String,
            required: true
        }
    },
    data() {
        return {
            localBucket: this.bucket
        }
    },
    methods: {
        openBucket() {
            let currentPath = this.$route.path.replace('/storage/', '');
            this.$router.push(`${currentPath}/${this.localBucket}`);
            this.$router.go(1);
        }
    }
}
</script>

<template>
    <div class="folder" @click="openBucket()">
        <div class="folder-header">
            <div class="folder-icon"></div>
            <p class="folder-title">{{this.localBucket}}</p>
        </div>
    </div>
</template>

<style>
.folder {
    border: 1px solid #ccc;
    width: 200px;
    padding: 10px;
    margin-bottom: 10px;
    border-radius: 5px;
    cursor: pointer;
    margin-right: 10px;
}

.folder-header {
    font-family: Arial, sans-serif;
    display: flex;
    align-items: center;
}
.folder-icon {
    width: 20px;
    height: 20px;
    background-color: lightsteelblue;
    border-radius: 3px;
    margin-right: 10px;
}
</style>

