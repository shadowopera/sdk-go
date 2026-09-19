`AtlasExtension` holds your own additions to the config data: computed values, config entry indexes, anything you want derived at load time.

Its source file is generated once and never regenerated, so it is safe to edit. Add your fields to the struct and build them in `AtlasExtension.OnLoaded`.

`AtlasExtension` is embedded in [ConfigAtlas](../atlas/). [ConfigAtlas.OnLoaded](../atlas/#onloaded) calls `AtlasExtension.OnLoaded` after every config table is loaded and every cross-table reference is bound.
