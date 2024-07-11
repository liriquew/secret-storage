<script>
import BucketItem from "@/components/BucketItem.vue"
export default {
  components: {BucketItem},
  props: {
    buckets: {
      type: Array,
      required: true,
    }
  }
}
</script>

<template>
    <div class="bucket-list"> 
      <BucketItem v-for="(bucket, index) in buckets" :key="index" :bucket="bucket" />
    </div>
</template>

<style>

.bucket-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(215px, 1fr));
  gap: 15px;
}

</style>

