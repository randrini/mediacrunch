<!--
  THESIS: A darkroom where every compression job is an exposure decision —
  quality is exposure time, progress is the enlarger countdown, the media
  grid is a contact sheet. Refuses the category default of true-black-plus-
  neon-accent SaaS dashboards.

  OWN-WORLD: Warm near-black ground (#0d0b0a, #16120e, #1f1a14, #2a2218)
  with safelight-red (#c8412a) as the single accent illuminating action and
  data; warm paper-white text (#f5f0e8) on a grayscale tonal ramp; Saira
  sans for UI, JetBrains Mono for data; ruled monospace data lines; contact-
  sheet grids with frame-number density; enlarger-bar progress fills with
  red glow; safelight-glow on active states.

  STORY: The visitor sees their media library as a contact sheet — every
  image's size, role, and compression state visible at a glance. They start
  a compression job and watch the enlarger countdown. They trust the tool
  because it shows exactly what was compressed, what was skipped, and why.

  FIRST VIEWPORT: Dashboard shows total savings as a developing print, three
  instance cards as enlarger stations with per-station stats, and recent
  compression jobs as a printing log. The nav carries an aperture icon in
  safelight red. Stats use the grayscale tonal ramp for size badges.

  FORM: The Darkroom, position 6 on the grounded list, raised by the
  split-flap board's live-row cascade and the ASCII render's monospace grid
  discipline. Seed key 4c56388d.

  FINISH: unreviewed and undocumented is unfinished; this build ends with
  the finish review, the verdict, DESIGN.md, and every shipping raster
  carrying its provenance.
-->
<template>
  <Layout>
    <router-view />
  </Layout>
</template>

<script setup lang="ts">
import Layout from './components/Layout.vue'
</script>
