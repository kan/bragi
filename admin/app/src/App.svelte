<script lang="ts">
  import { onMount } from "svelte";

  interface Config {
    port: string;
    admin_port: string;
    use_ai: boolean;
    use_lisp: boolean;
    year_format: string;
    month_format: string;
    date_format: string;
    date_time_format: string;
    time_zone: string;
    dictionary: Array<string> | null;
    dict_path: string;
  };

  let config: Config = {
    port: "", admin_port: "",
    use_ai: true, use_lisp: true,
    year_format: "", month_format: "", date_format: "", date_time_format: "",
    time_zone: "Asia/Tokyo", dictionary: null, dict_path: "",
  };
  let dicts:Array<string> = [];

  async function fetchData() {
    try {
      const res = await fetch('/api/config');
      if (res.ok) {
        config = await res.json();
        if (config.dictionary) {
          dicts = config.dictionary;
        }
      } else {
        console.error('fail API request');
      }
    } catch (err) {
      console.error('fail API request:', err);
    }
  }

  function addDict() {
    dicts = [...dicts, ''];
  }

  function removeDict(idx: number) {
    dicts = dicts.filter((_, i) => i != idx);
  }

  $: config.dictionary = dicts;

  let isSaving: boolean = false;

  async function saveConfig() {
    isSaving = true;

    await fetch('/api/config', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(config),
    });

    isSaving = false;
  }

  onMount(() => {
    fetchData();
  });
</script>

<main>
  <h1>Bragi</h1>

  <form>
    <fieldset>
      <label>
        SKKサーバーポート
        <input type="text" placeholder="1234" bind:value={config.port} />
      </label>
      <label>
        管理画面ポート
        <input type="text" placeholder="8080" bind:value={config.admin_port} disabled />
      </label>
      <label>
        <input type="checkbox" bind:checked={config.use_ai} />
        <span>AI辞書の使用</span>
      </label>
      <label>
        <input type="checkbox" bind:checked={config.use_lisp} />
        <span>Lisp辞書の使用</span>
      </label>
      <label>
        年の表記
        <input type="text" placeholder="2006年" bind:value={config.year_format} />
      </label>
      <label>
        月の表記
        <input type="text" placeholder="2006年1月" bind:value={config.month_format} />
      </label>
      <label>
        日の表記
        <input type="text" placeholder="2006年1月2日" bind:value={config.date_format} />
      </label>
      <label>
        時刻の表記
        <input type="text" placeholder="2006年1月2日 15時4分" bind:value={config.date_time_format} />
      </label>
      <label>
        タイムゾーン
        <input type="text" placeholder="UTC" bind:value={config.time_zone} />
      </label>
      <label>
        辞書ファイル
        {#each dicts as _, index}
        <fieldset role="group">
          <input type="text" bind:value={dicts[index]} />
          <input type="button" class="secondary" on:click={() => removeDict(index)} value="削除" />
        </fieldset>
        {/each}
        <div>
          <button type="button" on:click={addDict}>追加</button>
        </div>
      </label>
      <label>
        辞書ファイル保存場所
        <input type="text" placeholder="" bind:value={config.dict_path} />
      </label>
    </fieldset>
    <button type="button" on:click={saveConfig} disabled={isSaving}>
      {#if isSaving}
        保存中...
      {:else}
        保存
      {/if}
    </button>
  </form>
</main>

<style>
</style>
