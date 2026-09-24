# Apple Pay domain verification

Place the exact `apple-developer-merchantid-domain-association` file downloaded
from Stripe Dashboard in this directory before the production build. Do not
edit or generate its contents: Apple/Stripe validate the signed payload.

The deployed URL must be:

`https://<every-production-domain>/.well-known/apple-developer-merchantid-domain-association`
