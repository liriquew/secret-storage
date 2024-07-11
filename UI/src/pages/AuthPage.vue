<script>
import AuthField from "@/components/AuthField.vue";
import HeaderTitle from "@/components/HeaderTitle.vue"
export default {
  components: {
    AuthField,
    HeaderTitle,
  }
}
</script>

<template>
  <HeaderTitle/>
  <div class="content"> 
    <AuthField/>
  </div>  
</template>

<style>
.content {
  padding-top: 100px;
}
</style>