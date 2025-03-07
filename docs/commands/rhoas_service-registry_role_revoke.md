## rhoas service-registry role revoke

Revoke role for principal

### Synopsis

Revoke the role of a user or service account.

NOTE: It is not possible to revoke the role of the owner of the instance. The instance owner always retains ADMIN rights.


```
rhoas service-registry role revoke [flags]
```

### Examples

```
## Revoke role for user
rhoas service-registry role revoke --username=janedough

```

### Options

```
      --instance-id string       ID of the Service Registry instance to be used (by default, uses the currently selected instance)
      --service-account string   ServiceAccount name
      --username string          Username of the user within organization
```

### Options inherited from parent commands

```
  -h, --help      Show help for a command
  -v, --verbose   Enable verbose mode
```

### SEE ALSO

* [rhoas service-registry role](rhoas_service-registry_role.md)	 - Service Registry role management

