# Go app using docker with postgres

## load the env files to use in the app

```bash
source ./load_env.sh
```

## to run seeds

```bash
goose -dir ./data/seeds -no-versioning up
```

# Todo

Need to look at the way that time is shown on the uploads page

## Vist creation

The handler that handles the visit. Can remove the validation step there.

The idea here is to rejig the html form into a 2 panel toggle.

The first button the users presses validates the form and if valid hides the fields in a display none div.
A preview / review panel is then toggled in with the values in - written in by js - where the user can see their "input" before they submit.

In this review panel will be two buttons, a button that toggles back to the form panel so they can make adustments - and a submit button. T

he submit button will call the bandler as is more or less. Any validation added on the handler, possible....?

The handler now has to be rejigged.

1). We'll need a writable directory on the host to write the images to.
2). Well need a process of tying all the cross table data together

use the last_insert_id eg.

```sql
INSERT INTO vist (names...) VALUES (values...);
```

This will return a last_insert_id that we use as the visit_id for each image added tothe images table.

So for each image ...

    create a filename hash
    save to disk
    insert into images storing the hash so we can ref it in the app
    maybe create a thumbnail at this point as well

Return some HTML to tell the user the visit is complete..

Spend about 3 hours today creating the file upload sql. Hit a few snags one of them was to do with not importing "image/jpeg" which is what allows go to image.DecodeConfig.

TODO -

Add those extra fields and create the thumbs on the preview
Add pico CSS
Wrap in a PWA

GET visit/pick-a-customer
POST visit/pick-a-location
POST visit/log-a-location
visit/confirm-visit

Research pico css

https://picocss.com/docs/conditional
duration slider needs a look

images thumbs
size and formatting
