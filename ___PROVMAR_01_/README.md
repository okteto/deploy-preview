PROVMA — aperçu local

Ce dossier contient une petite vitrine statique (HTML/CSS/JS) pour le projet PROVMA.

Comment prévisualiser localement

1. Depuis le conteneur ou la machine contenant ce repo, démarrez un serveur HTTP simple :

```bash
python3 -m http.server 8000 --directory /workspaces/deploy-preview/___PROVMAR_01_
```

2. Ouvrez votre navigateur à : http://localhost:8000

Remarques
- Les fichiers médias originaux ont été normalisés (espaces et apostrophes remplacés) pour éviter les problèmes d'URL.
- Le lecteur vidéo a été converti en lecteur HTML5 et la playlist est interactive. Le premier élément est sélectionné automatiquement au chargement.

Options de déploiement public (rapide)
- GitHub Pages : poussez le dossier sur une branche `gh-pages` ou configurez GitHub Pages depuis la branche `main`.
- Netlify : glissez-déposez le dossier dans Netlify Drop ou configurez un site lié au repo.
- Docker : une image simple peut servir ce dossier avec `nginx` ou `http.server`.
- Tunnel public : utilisez `ngrok http 8000` si vous voulez une URL temporaire.

Si vous voulez, je peux :
- préparer un déploiement Netlify/GitHub Pages automatiquement;
- créer une image Docker qui sert le dossier et publier un preview;
- optimiser les vidéos volumineuses et générer miniatures automatiques.

Dites-moi ce que vous souhaitez que je fasse ensuite.