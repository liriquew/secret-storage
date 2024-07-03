<script>
import KVListItem from "@/components/KVListItem.vue";

export default {
  components: {KVListItem},
  props: {
    kvs: {
      type: Array,
      required: true,
    },
  }
}
</script>

<template>
  <KVListItem v-for="(kv, index) in kvs" :key="index" :kv="kv" />
</template>

<style scoped>
</style>

