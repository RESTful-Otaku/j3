<script module>
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import Pagination from '../../lib/components/Pagination.svelte';
  const { Story } = defineMeta({ title: 'J3/Pagination', component: Pagination, tags: ['autodocs'] });
</script>
<Story name="Default" args={{ total: 87, page: 1, pageSize: 10 }} />
<Story name="Empty" args={{ total: 0, page: 1, pageSize: 10 }} />
