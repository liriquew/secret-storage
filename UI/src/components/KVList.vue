<script>
import KVListItem from "@/components/KVListItem.vue";

export default {
  components: {KVListItem},
  props: {
    kvs: {
      type: Array,
      required: true,
    },
  },
  emits: ['deleteKV'],
  methods: {
    deleteKV(key) {
      this.$emit('deleteKV', key);
    },
  }
}
</script>

<template>
  <KVListItem @deleteKV="deleteKV" v-for="kv in kvs" :key="kv.key" :kv="kv" />
</template>

<style scoped>
</style>

