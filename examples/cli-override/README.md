# CLI Override

When a single YAML file is not suitable due to automation practices (e.g. GitOps), and want your files to remain untouch yet work fine... passing credentials can be hazardous.

Be reassure: not with CLI overrides :partying_face:

Consider the following `admin` config.
```yaml
admin:
  name: 'placeholder'
  email: 'placeholder'
  password: 'placeholder'
```

You can run `ctfd-setup` using the following, in order to override these settings (note that environment variables would have worked too).
```bash
ctfd-setup \
    --url http://localhost:8000 \
    --file .ctfd.yaml \
    --admin.name="ctfer" \
    --admin.email="ctfer-io@protonmail.com" \
    --admin.password="ctfer"
```

Now, you can log in user `ctfer` account (same as password), and not `placeholder`!
