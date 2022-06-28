# cdnvalidator service

The cdnvalidator provides an abstraction layer for Cloudfront invalidations. It delegates access for different owners of distribution paths to manage performance and cache control requirements.

## Architecture

The CDN Validator provides a RESTful API detailed in the [OpenAPI documentation](./swagger/swagger.json).

### Authn/Authz

The service relies on being deployed behind a reverse proxy that handles forward authentication and validation of the JWT.  Authorization decisions will be made against the `groups` and `scp` claims of a JWT.

## Configuration

```yaml
distributions:
    sandbox:
        id: "<Cloudfront Distribution ID>"
        prefix: "/my/path"
entitlements:
    mygroup:
    - sandbox
```

The `sandbox` is the vanity name representing a virtual distribution along the path prefix `/my/path`.  The entitlement `mygroup` will only be allowed to submit invalidation requests for `/my/path/*` resources.

* Many vanity names MAY be created with the same Cloudfront distribution ID
* Entitlements MAY be assigned to more than one distribution.
* Vanity distributions MUST not conflict in paths. 
