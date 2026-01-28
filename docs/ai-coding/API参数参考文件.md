# API搜索参数

## Enter search term:

```html
<input type="text" class="input" name="q" size="64" maxlength="255">
```

## Select search type:

```html
<select name="searchtype" size="19">
            <option value="c">CERTIFICATE</option>
            <option value="id">&nbsp; crt.sh ID</option>
            <option value="ctid">&nbsp; CT Entry ID</option>
            <option value="serial">&nbsp; Serial Number</option>
            <option value="ski">&nbsp; Subject Key Identifier</option>
            <option value="spkisha1">&nbsp; SHA-1(SubjectPublicKeyInfo)</option>
            <option value="spkisha256">&nbsp; SHA-256(SubjectPublicKeyInfo)</option>
            <option value="subjectsha1">&nbsp; SHA-1(Subject)</option>
            <option value="sha1">&nbsp; SHA-1(Certificate)</option>
            <option value="sha256">&nbsp; SHA-256(Certificate)</option>
            <option value="ca">CA</option>
            <option value="CAID">&nbsp; ID</option>
            <option value="CAName">&nbsp; Name</option>
            <option value="Identity" selected="">IDENTITY</option>
            <option value="CN">&nbsp; commonName (Subject)</option>
            <option value="E">&nbsp; emailAddress (Subject)</option>
            <option value="OU">&nbsp; organizationalUnitName (Subject)</option>
            <option value="O">&nbsp; organizationName (Subject)</option>
            <option value="dNSName">&nbsp; dNSName (SAN)</option>
            <option value="rfc822Name">&nbsp; rfc822Name (SAN)</option>
            <option value="iPAddress">&nbsp; iPAddress (SAN)</option>
          </select>
```

## Select search options:

```html
<div style="border:1px solid #AAAAAA;margin-bottom:5px;padding:4px 2px;text-align:left">
            &nbsp;<select name="match">
              <option value="" selected="">Autoselect</option>
              <option value="=">=</option>
              <option value="ILIKE">ILIKE</option>
              <option value="LIKE">LIKE</option>
              <option value="single">Single</option>
              <option value="any">Any</option>
              <option value="FTS">Full Text Search</option>
            </select> Identity matching
            <br><input type="checkbox" name="excludeExpired"> Exclude expired certificates?
            <br><input type="checkbox" name="deduplicate"> Deduplicate (pre)certificate pairs?
            <br><input type="checkbox" name="showSQL"> Show SQL?
            <hr>
            &nbsp;Or, <input type="checkbox" name="searchCensys"> Search on <span style="vertical-align:-30%"><img src="/censys.png"></span>?
          </div>
```

## Select linting options:

```html
<span class="heading">Select linting options:</span>
<select name="linter" size="3">
            <option value="cablint">cablint</option>
            <option value="x509lint">x509lint</option>
            <option value="zlint" selected="">zlint</option>
            <option value="keylint">keylint</option>
            <option value="lint">ALL</option>
          </select>
<select name="linttype" size="3">
            <option value="1 week" selected="">1-week Summary</option>
            <option value="issues">Issues</option>
          </select>
```

























